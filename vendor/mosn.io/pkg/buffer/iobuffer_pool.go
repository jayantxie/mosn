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

package buffer

import (
	"errors"
	"sync"
)

var ibPool IoBufferPool

// IoBufferPool is Iobuffer Pool
type IoBufferPool struct {
	pool sync.Pool
}

// take returns IoBuffer from IoBufferPool
func (p *IoBufferPool) take(size int) (buf IoBuffer) {
	v := p.pool.Get()
	if v == nil {
		buf = newIoBuffer(size)
	} else {
		buf = v.(IoBuffer)
		buf.Alloc(size)
		buf.Count(1)
	}
	return
}

// give returns IoBuffer to IoBufferPool
func (p *IoBufferPool) give(buf IoBuffer) {
	buf.Free()
	p.pool.Put(buf)
}

// GetIoBuffer returns IoBuffer from pool
func (p *IoBufferPool) GetIoBuffer(size int) IoBuffer {
	return p.take(size)
}

// PutIoBuffer returns IoBuffer to pool
func (p *IoBufferPool) PutIoBuffer(buf IoBuffer) error {
	count := buf.Count(-1)
	if count > 0 {
		return nil
	} else if count < 0 {
		return errors.New("PutIoBuffer duplicate")
	}

	if pb, _ := buf.(*pipe); pb != nil {
		buf = pb.IoBuffer
	}
	p.give(buf)
	return nil
}

// GetIoBuffer is a wrapper for ibPool
func GetIoBuffer(size int) IoBuffer {
	return ibPool.GetIoBuffer(size)
}

// NewIoBuffer is an alias for GetIoBuffer
func NewIoBuffer(size int) IoBuffer {
	return ibPool.GetIoBuffer(size)
}

// PutIoBuffer is a a wrapper for ibPool
func PutIoBuffer(buf IoBuffer) error {
	return ibPool.PutIoBuffer(buf)
}
