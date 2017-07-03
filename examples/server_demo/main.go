package main

import (
	"fmt"
	. "github.com/domac/husky"
	"github.com/domac/husky/log"
	"github.com/domac/husky/pb"
)

func main() {
	log.GetLogger().Infoln("Start Server")

	simpleServer := NewServer("localhost:10028", nil, func(remoteClient *HuskyClient, p *Packet) {

		if p.Header.ContentType == pb.CMD_BYTES_MESSAGE {
			bm := &pb.BytesMessage{}
			pb.UnmarshalPbMessage(p.Data, bm)

			testData := bm.GetBody()
			//直接回写回去
			req := string(testData)
			resp := NewPbBytesPacket(p.Header.PacketId, "demoserver", []byte(req+"_resp"))
			remoteClient.Write(*resp)
		} else {
			fmt.Println("get string")
		}

	})
	simpleServer.ListenAndServer()
	<-simpleServer.StopChan
}
