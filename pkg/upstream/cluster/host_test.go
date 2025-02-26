/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cluster

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
	"mosn.io/api"
	v2 "mosn.io/mosn/pkg/config/v2"
	"mosn.io/pkg/buffer"
	"mosn.io/pkg/utils"
)

// TestCreateConnectionIdleTimeout test the connection idle timeout feature
func TestCreateConnectionIdleTimeout(t *testing.T) {
	assert := assert.New(t)

	// create host
	ctx := context.Background()
	clusterConf := v2.Cluster{
		Name:        "mock",
		ClusterType: v2.SIMPLE_CLUSTER,
		LbType:      v2.LB_ROUNDROBIN,
		Hosts: []v2.Host{
			{
				HostConfig: v2.HostConfig{
					Address: "127.0.0.1:10086",
				},
			},
		},
		IdleTimeout: &api.DurationConfig{
			Duration: 2 * time.Second,
		},
	}

	host := NewSimpleHost(clusterConf.Hosts[0], NewCluster(clusterConf).Snapshot().ClusterInfo())
	assert.NotNil(host)

	// start mock server
	server := &countConnServer{
		address: "127.0.0.1:10086",
	}

	server.start()
	defer server.stop()

	// reset read timeout
	oldConnReadTimeout := buffer.ConnReadTimeout
	buffer.ConnReadTimeout = 300 * time.Millisecond
	defer func() {
		buffer.ConnReadTimeout = oldConnReadTimeout
	}()

	// create connection
	checkCountOne := func() bool {
		count := server.getCount()
		return count == 1
	}
	conn := host.CreateConnection(ctx)
	conn.Connection.Connect()
	conn.Connection.Start(ctx)

	// we create a connection to port 10086, so there should be only one connection on server
	assert.Eventually(checkCountOne, 1*time.Second, 500*time.Millisecond, "expect one connection")

	checkCountZero := func() bool {
		count := server.getCount()
		return count == 0
	}
	// the connection we created should be closed by idle timeout checker
	// so there should be no connection on server
	assert.Eventually(checkCountZero, 3*time.Second, 500*time.Millisecond, "expect non connection")
}

type countConnServer struct {
	address  string
	count    atomic.Int64
	listener net.Listener
	stopChan chan struct{}
}

func newCountConnServer(address string) *countConnServer {
	return &countConnServer{
		address:  address,
		stopChan: make(chan struct{}),
	}
}

func (c *countConnServer) start() error {
	c.stopChan = make(chan struct{})
	listener, err := net.Listen("tcp4", c.address)
	if err != nil {
		return err
	}
	c.listener = listener

	utils.GoWithRecover(func() {
		for {
			conn, err := c.listener.Accept()
			if err != nil {
				break
			}
			c.count.Inc()
			utils.GoWithRecover(func() {
				defer c.count.Dec()
				for {
					conn.SetReadDeadline(time.Now().Add(time.Second * 15))

					var buf = make([]byte, 1024)
					_, err := conn.Read(buf)
					if err != nil {
						return
					}

					select {
					case <-c.stopChan:
						return
					default:
					}
				}
			}, nil)
		}
	}, nil)

	return nil
}

func (c *countConnServer) getCount() int64 {
	return c.count.Load()
}

func (c *countConnServer) stop() {
	c.listener.Close()
	close(c.stopChan)
}
