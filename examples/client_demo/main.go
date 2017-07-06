package main

import (
	"context"
	"fmt"
	. "github.com/domac/husky"
	"github.com/domac/husky/pb"
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
		bm := &pb.BytesMessage{}
		pb.UnmarshalPbMessage(resp.([]byte), bm)
		fmt.Printf("resp >=====> %s|%s\n", string(bm.GetBody()), bm.GetHeader().GetMessageType())
	}
	<-ctx.Done()
	println("client Down !!")
	cancel()
}
