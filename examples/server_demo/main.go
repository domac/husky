package main

import (
	. "github.com/domac/husky"
	"github.com/domac/husky/log"
	"github.com/domac/husky/pb"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	log.GetLogger().Infoln("Start Server")

	simpleServer := NewServer("localhost:10028", nil, func(remoteClient *HuskyClient, p *Packet) {

		if p.Header.ContentType == PB_BYTES_MESSAGE {
			bm := &pb.BytesMessage{}
			UnmarshalPbMessage(p.Data, bm)

			testData := bm.GetBody()
			//直接回写回去
			req := string(testData)
			resp := NewPbBytesPacket(p.Header.PacketId, "demo_server_function", []byte(req+"_resp"))
			remoteClient.Write(*resp)
		} else {
			resp := NewPacket([]byte("get string"))
			remoteClient.Write(*resp)
		}

	})
	simpleServer.ListenAndServer()

	http.ListenAndServe(":9090", nil)

	<-simpleServer.StopChan
}
