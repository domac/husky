package husky

import (
	"bytes"
	"fmt"
	"github.com/domac/husky/pb"
	"github.com/golang/protobuf/proto"
	"time"
)

const (
	PACKET_HEADER_LENGTH = (4 + 1 + 2 + 4)
	MAX_BYTES            = 32 * 1024
)

type Packet struct {
	Header PacketHeader
	Data   []byte
}

func NewPacket(data []byte) *Packet {
	h := PacketHeader{
		PacketId:    -1,
		ContentType: uint8(0x01),
		BodyLenth:   int32(len(data))}

	return &Packet{
		Header: h,
		Data:   data}
}

func NewCmdPacket(packetId int32, cmdtype uint8, data []byte) *Packet {
	p := NewPacket(data)
	p.Header.PacketId = packetId
	p.Header.ContentType = cmdtype
	return p
}

func NewPbBytesPacket(packetId int32, messageType string, data []byte) *Packet {

	header := &pb.Header{
		MessageId:   proto.String(fmt.Sprintf("%d", packetId)),
		MessageType: proto.String(messageType),
		CreateTime:  proto.Int64(time.Now().Unix())}

	dataBytes := pb.MarshalMessage(header, pb.PB_BYTES_MESSAGE, data)

	p := NewCmdPacket(packetId, pb.PB_BYTES_MESSAGE, dataBytes)
	p.Header.PacketId = packetId
	return p
}

func NewPbStringPacket(packetId int32, messageType string, data string) *Packet {
	header := &pb.Header{
		MessageId:   proto.String(fmt.Sprintf("%d", packetId)),
		MessageType: proto.String(messageType),
		CreateTime:  proto.Int64(time.Now().Unix())}

	dataBytes := pb.MarshalMessage(header, pb.PB_STRING_MESSAGE, data)

	p := NewCmdPacket(packetId, pb.PB_STRING_MESSAGE, dataBytes)
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
