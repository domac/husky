Husky
=======
基础协议通讯服务层, 支持protobuf


## 获取相关依赖:

```
$ go get -u -v github.com/Sirupsen/logrus

$ go get -u -v github.com/golang/protobuf

```

## 使用参考:

### 服务端运行例子:

```go
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

		if p.Header.ContentType == pb.PB_BYTES_MESSAGE {
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
	"time"
)

func main() {
	conn, _ := Dial("localhost:10028")
	simpleClient := NewClient(conn, nil, nil)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 50*time.Second)
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

```