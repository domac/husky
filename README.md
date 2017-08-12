Husky
=======
基础协议通讯服务层, 支持protobuf


## 获取相关依赖:

```
$ go get -u -v github.com/golang/protobuf
```

## 使用参考:

### 服务端运行例子:

```go
package main

import (
	. "github.com/domac/husky"
	"github.com/domac/husky/pb"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func main() {

	rateLimitNum := 5000 //限流速率
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

```

### 客户端调用例子

```go
package main

import (
	"fmt"
	. "github.com/domac/husky"
	"github.com/domac/husky/pb"
	"time"
)

func main() {
	conn, _ := Dial("localhost:10028")
	simpleClient := NewClient(conn, nil, nil)
	simpleClient.Start()

	for i := 0; i < 1000; i++ {
		p := NewPbBytesPacket(1, "democlient", []byte("husky"))
		resp, _ := simpleClient.SyncWrite(*p, 500*time.Millisecond)
		bm := &pb.BytesMessage{}
		pb.UnmarshalPbMessage(resp.([]byte), bm)
		fmt.Printf("resp >=====> %s|%s\n", string(bm.GetBody()), bm.GetHeader().GetMessageType())
	}
	simpleClient.Shutdown()
}
```

### 客户端带超时context例子

```go
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

```