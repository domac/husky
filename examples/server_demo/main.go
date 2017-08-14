package main

import (
	. "github.com/domac/husky"
	"github.com/domac/husky/pb"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func main() {

	rateLimitNum := 8000 //限流速率
	cfg := NewConfig(1000, 4*1024, 4*1024, 10000, 10000, 10*time.Second, 160000, -1, rateLimitNum)

	simpleServer := NewServer("localhost:10028", cfg, func(remoteClient *HClient, p *Packet) {

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

	//开启http, 用于debug
	http.ListenAndServe(":9090", nil)

	<-simpleServer.StopChan
}
