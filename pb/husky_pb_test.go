package pb

import (
	"github.com/golang/protobuf/proto"
	"testing"
	"time"
)

func TestPb(t *testing.T) {
	bodyData := "test simple husky"

	p := &StringMessage{
		Body: proto.String(bodyData),
		Header: &Header{
			MessageId:   proto.String("2017"),
			MessageType: proto.String("normal"),
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

func TestHuskyBytesMessage(t *testing.T) {
	data := []byte("husky test bytes messages")

	header := &Header{
		MessageId:   proto.String("temp-001"),
		MessageType: proto.String("test bytes"),
		CreateTime:  proto.Int64(time.Now().Unix())}

	dataBytes := MarshalMessage(header, CMD_BYTES_MESSAGE, data)
	//bm := buildBytesMessage(data)

	tempBytesMessage := &BytesMessage{}
	UnmarshalPbMessage(dataBytes, tempBytesMessage)
	println(string(tempBytesMessage.GetBody()))
	println(tempBytesMessage.GetHeader().GetMessageId())
}

func TestHuskyStringMessage(t *testing.T) {
	data := "husky string bytes messages"

	header := &Header{
		MessageId:   proto.String("temp-002"),
		MessageType: proto.String("test string"),
		CreateTime:  proto.Int64(time.Now().Unix())}

	dataBytes := MarshalMessage(header, CMD_STRING_MESSAGE, data)
	//bm := buildBytesMessage(data)

	tempBytesMessage := &StringMessage{}
	UnmarshalPbMessage(dataBytes, tempBytesMessage)
	println(string(tempBytesMessage.GetBody()))
	println(tempBytesMessage.GetHeader().GetMessageId())
}
