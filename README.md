Husky
=======
基于protobuf协议的CS服务层

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