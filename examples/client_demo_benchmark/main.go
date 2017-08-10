package main

import (
	"context"
	"fmt"
	. "github.com/domac/husky"
	"github.com/domac/husky/pb"
	"sync/atomic"
	"time"
)

//go tool pprof --seconds 50 http://localhost:9090/debug/pprof/profile

func main() {
	conn, _ := Dial("localhost:10028")
	simpleClient := NewClient(conn, nil, nil)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 100*time.Second)
	defer cancel()

	simpleClient.StartWithContext(ctx)

	count := int64(0)

	for !simpleClient.IsClosed() {

		atomic.AddInt64(&count, 1)

		select {
		case <-ctx.Done():
			goto FINISH
		default:
			p := NewPbBytesPacket(1, "democlient", []byte("husky"))
			resp, _ := simpleClient.SyncWrite(*p, 25*time.Millisecond)
			bm := &pb.BytesMessage{}
			if resp != nil {
				UnmarshalPbMessage(resp.([]byte), bm)
				fmt.Printf("benchmark resp >=====> %s|%s|%d\n", string(bm.GetBody()), bm.GetHeader().GetFunctionType(), count)
			}
		}
	}

FINISH:
	println("client Down !!")
	cancel()
}
