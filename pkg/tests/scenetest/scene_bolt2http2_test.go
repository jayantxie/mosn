package tests

import (
	"encoding/binary"
	"encoding/hex"
	"testing"
	"time"

	"github.com/orcaman/concurrent-map"
	"gitlab.alipay-inc.com/afe/mosn/pkg/mosn"
	"gitlab.alipay-inc.com/afe/mosn/pkg/protocol"
)

func TestBolt2Http2(t *testing.T) {
	http2Addr := "127.0.0.1:8080"
	meshAddr := "127.0.0.1:2045"
	server := NewUpstreamHttp2(t, http2Addr)
	server.GoServe()
	defer server.Close()
	mesh_config := CreateSimpleMeshConfig(meshAddr, []string{http2Addr}, protocol.SofaRpc, protocol.Http2)
	go mosn.Start(mesh_config, "", "")
	time.Sleep(5 * time.Second) //wait mesh and server start
	boltV1ReqBytes, _ := hex.DecodeString("0101000101000000010100001388002c002e000005b5636f6d2e616c697061792e736f66612e7270632e636f72652e726571756573742e536f66615265717565737400000007736572766963650000001f636f6d2e616c697061792e746573742e54657374536572766963653a312e304fbc636f6d2e616c697061792e736f66612e7270632e636f72652e726571756573742e536f666152657175657374950d7461726765744170704e616d650a6d6574686f644e616d651774617267657453657276696365556e697175654e616d650c7265717565737450726f70730d6d6574686f64417267536967736f904e076563686f5374721f636f6d2e616c697061792e746573742e54657374536572766963653a312e304d03617070037878780870726f746f636f6c04626f6c74117270635f74726163655f636f6e746578744d09736f66615270634964013007456c61737469634e0b73797350656e4174747273000d736f666143616c6c657249646300097a70726f78795549444e107a70726f78795461726765745a6f6e654e0c736f666143616c6c65724970000b736f6661547261636549641d30613066653865663135323431343435383331373231303031393836300c736f666150656e4174747273000e736f666143616c6c65725a6f6e654e097a70726f78795669704e0d736f666143616c6c6572417070037878787a7a567400075b737472696e676e01106a6176612e6c616e672e537472696e677a53040031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334353637383930313233343536373839303132333435363738393031323334")

	client := &RpcClient{
		t:               t,
		addr:            meshAddr,
		response_filter: &Http2Response{},
		wait_reponse:    cmap.New(),
	}
	if err := client.Connect(); err != nil {
		t.Fatalf("client connect failed\n")
	}
	var Id uint32 = 1
	for ; Id <= 20; Id++ {
		requestIdBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(requestIdBytes, Id)
		copy(boltV1ReqBytes[5:], requestIdBytes)
		client.SendRequest(Id, boltV1ReqBytes)
	}
	//client.wait_response should empty
	<-time.After(10 * time.Second)
	if !client.wait_reponse.IsEmpty() {
		t.Errorf("exists request no response\n")
		t.Logf("%v\n", client.wait_reponse.Keys())
	}
}
