package husky

import (
	"bytes"
	"fmt"
	"github.com/domac/husky/pb"
	"github.com/golang/protobuf/proto"
	"log"
	"time"
)

const (
	PACKET_HEADER_LENGTH = (4 + 1 + 2 + 4)

	//message
	NORMAL_MESSAGE    = uint8(0x01) //普通类型
	ERROR_MESSAGE     = uint8(0x02) //错误类型
	PB_BYTES_MESSAGE  = uint8(0x03) //protobuff bytes类型
	PB_STRING_MESSAGE = uint8(0x04) //protobuff string类型
)

type Packet struct {
	Header PacketHeader
	Data   []byte
}

//创建数据包
func NewPacket(data []byte) *Packet {
	h := PacketHeader{
		PacketId:    0,
		ContentType: NORMAL_MESSAGE,
		BodyLenth:   int32(len(data))}

	return &Packet{
		Header: h,
		Data:   data}
}

//定制数据包
func buildPacket(packetId int32, cmdtype uint8, data []byte) *Packet {
	p := NewPacket(data)
	p.Header.PacketId = packetId
	p.Header.ContentType = cmdtype
	return p
}

func NewPbBytesPacket(packetId int32, functionType string, data []byte) *Packet {

	header := &pb.Header{
		MessageId:    proto.String(fmt.Sprintf("%d", packetId)),
		FunctionType: proto.String(functionType),
		CreateTime:   proto.Int64(time.Now().Unix())}

	dataBytes := GetMarshalMessage(header, PB_BYTES_MESSAGE, data)

	p := buildPacket(packetId, PB_BYTES_MESSAGE, dataBytes)
	p.Header.PacketId = packetId
	return p
}

func NewPbStringPacket(packetId int32, functionType string, data string) *Packet {
	header := &pb.Header{
		MessageId:    proto.String(fmt.Sprintf("%d", packetId)),
		FunctionType: proto.String(functionType),
		CreateTime:   proto.Int64(time.Now().Unix())}

	dataBytes := GetMarshalMessage(header, PB_STRING_MESSAGE, data)

	p := buildPacket(packetId, PB_STRING_MESSAGE, dataBytes)
	p.Header.PacketId = packetId
	return p
}

func (p *Packet) MarshalPacket() []byte {
	dl := 0
	if nil != p.Data {
		dl = len(p.Data)
	}
	buff := MarshalHeader(p.Header, int32(dl))
	buff.Write(p.Data)
	return buff.Bytes()
}

func (p *Packet) ResetPacket() {
	p.Header.PacketId = -1
}

type PacketHeader struct {
	PacketId    int32 //请求序列编号
	ContentType uint8 //文本类型
	Version     int16 //协议版本号
	BodyLenth   int32 //包体长度
}

//序列化包头
func MarshalHeader(header PacketHeader, packetBodyLength int32) *bytes.Buffer {

	sb := make([]byte, 0, 4+PACKET_HEADER_LENGTH+packetBodyLength)
	buff := bytes.NewBuffer(sb)

	Write(buff, BIG_BYTE_ORDER, int32(PACKET_HEADER_LENGTH+packetBodyLength)) // 写4字节
	Write(buff, BIG_BYTE_ORDER, header.PacketId)
	Write(buff, BIG_BYTE_ORDER, header.ContentType)
	Write(buff, BIG_BYTE_ORDER, header.Version)
	Write(buff, BIG_BYTE_ORDER, header.BodyLenth)

	return buff
}

func UnmarshalHeader(r *bytes.Reader) (PacketHeader, error) {
	header := PacketHeader{}

	err := Read(r, BIG_BYTE_ORDER, &(header.PacketId))
	if err != nil {
		return header, nil
	}

	err = Read(r, BIG_BYTE_ORDER, &(header.ContentType))
	if err != nil {
		return header, nil
	}

	err = Read(r, BIG_BYTE_ORDER, &(header.Version))
	if err != nil {
		return header, nil
	}

	err = Read(r, BIG_BYTE_ORDER, &(header.BodyLenth))
	if err != nil {
		return header, nil
	}

	return header, nil
}

// --------  处理prorobuf消息

// 序列化消息
func GetMarshalMessage(header *pb.Header, msgType uint8, body interface{}) []byte {
	switch msgType {
	case PB_BYTES_MESSAGE:
		message := &pb.BytesMessage{}
		message.Header = header
		message.Body = body.([]byte)

		data, err := proto.Marshal(message)
		if nil != err {
			log.Fatalf("Marshall Bytes Message Error |%s|%d|%s\n", header, msgType, err)
		}
		return data
	case PB_STRING_MESSAGE:
		message := &pb.StringMessage{}
		message.Header = header
		message.Body = proto.String(body.(string))
		data, err := proto.Marshal(message)
		if nil != err {
			log.Fatalf("Marshall String Message Error |%s|%d|%s\n", header, msgType, err)
		}
		return data
	}
	return nil
}

func UnmarshalPbMessage(data []byte, msg proto.Message) error {
	return proto.Unmarshal(data, msg)
}

func MarshalPbString(s string) *string {
	return proto.String(s)
}

func MarshalInt32(i int32) *int32 {
	return proto.Int32(i)
}

func MarshalInt64(i int64) *int64 {
	return proto.Int64(i)
}

func MarshalPbMessage(message proto.Message) ([]byte, error) {
	return proto.Marshal(message)
}
