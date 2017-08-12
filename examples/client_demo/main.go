package main

import (
	"context"
	"fmt"
	. "github.com/domac/husky"
	"github.com/domac/husky/pb"
	_ "net/http/pprof"
	"time"
)

func main() {
	conn, _ := Dial("localhost:10028")
	simpleClient := NewClient(conn, nil, nil)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	simpleClient.StartWithContext(ctx)

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

	<-ctx.Done()
	println("client Down !!")
	cancel()
}
