package husky

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
)

//Husky的解码器
type HuskyCodec struct {
	MaxLength int32 //传输包的最大幅度
}

func (self *HuskyCodec) Read(reader *bufio.Reader) (*bytes.Buffer, error) {
	//总长度
	var length int32
	err := Read(reader, BIG_BYTE_ORDER, &length)
	if nil != err {
		return nil, err
	} else if length <= 0 {
		return nil, errors.New("TOO SHORT PACKET")
	}

	//总长度校验
	if length > self.MaxLength {
		return nil, errors.New("TOO LARGE PACKET")
	}

	buff := make([]byte, int(length))
	tmp := buff
	l := 0
	for {
		rl, err := reader.Read(tmp)
		if nil != err {
			return nil, err
		}
		l += rl

		if l < int(length) {
			tmp = tmp[rl:]
			continue
		} else {
			break
		}
	}
	return bytes.NewBuffer(buff), nil
}

//序列化
func (self *HuskyCodec) MarshalPacket(p *Packet) []byte {
	return p.MarshalPacket()
}

//反序列化
func (self *HuskyCodec) UnmarshalPacket(buff *bytes.Buffer) (*Packet, error) {
	p := &Packet{}

	if buff.Len() < PACKET_HEADER_LENGTH {
		return nil, errors.New(
			fmt.Sprintf("packet is less than limit length:%d/%d", buff.Len(), PACKET_HEADER_LENGTH))
	}
	reader := bytes.NewReader(buff.Next(PACKET_HEADER_LENGTH))
	//构造包头
	header, err := UnmarshalHeader(reader)
	if nil != err {
		return nil, errors.New(
			fmt.Sprintf("error unmarshaller header : %s", err.Error()))
	}

	p.Header = header
	p.Data = buff.Bytes()
	return p, nil
}
