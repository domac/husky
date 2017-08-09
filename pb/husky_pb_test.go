package pb

import (
	"github.com/golang/protobuf/proto"
	"testing"
)

func TestPb(t *testing.T) {
	bodyData := "test simple husky"

	p := &StringMessage{
		Body: proto.String(bodyData),
		Header: &Header{
			MessageId:    proto.String("2017"),
			FunctionType: proto.String("normal"),
		},
	}

	pData, err := proto.Marshal(p)

	if err != nil {
		println(err.Error())
	}
	println(string(pData))

	p2 := &StringMessage{}
	proto.Unmarshal(pData, p2)

	println(p2.GetHeader().GetMessageId())
}
