package main

import (
	"fmt"
	. "github.com/domac/husky"
	"github.com/domac/husky/pb"
	_ "net/http/pprof"
	"time"
)

func main() {
	conn, _ := Dial("localhost:10028")
	cfg := NewConfig(1000, 4*1024, 4*1024, 10000, 10000, 5*time.Second, 160000, -1, 1000, 0)
	simpleClient := NewClient(conn, cfg, nil)

	simpleClient.Start()

	for i := 0; i < 1000; i++ {
		p := NewPbBytesPacket(1, "democlient", []byte("husky"))
		resp, _ := simpleClient.SyncWrite(*p, 5*time.Millisecond)

		if resp == nil {
			continue
		}

		bm := &pb.BytesMessage{}
		UnmarshalPbMessage(resp.([]byte), bm)
		fmt.Printf("resp >=====> %s|%s\n", string(bm.GetBody()), bm.GetHeader().GetFunctionType())
	}

	for j := 0; j < 5; j++ {
		p := NewPacket([]byte("123"))
		resp, _ := simpleClient.SyncWrite(*p, 15*time.Millisecond)
		if resp != nil {
			fmt.Printf("%v\n", string(resp.([]byte)))
		}
	}

	simpleClient.Shutdown()

	p := NewPacket([]byte("123"))
	_, err := simpleClient.SyncWrite(*p, 1*time.Second)
	if err != nil {
		println(err.Error())
	}

	if simpleClient.IsClosed() {
		println("======== simpleClient is closed")
		b, err := simpleClient.Reconnect()
		if err != nil {
			println("Reconnect error")
			return
		}

		if !b {
			println("Reconnect fail")
			return
		} else {
			println("======== Reconnect success")

			for j := 0; j < 5; j++ {
				p := NewPacket([]byte("123"))
				resp, _ := simpleClient.SyncWrite(*p, 15*time.Millisecond)
				if resp != nil {
					fmt.Printf(">>>>>> %v\n", string(resp.([]byte)))
				}
			}
		}
	}

	println("client Down !!")
	select {}
}
