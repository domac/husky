package pb

import (
	"github.com/domac/husky/log"
	"github.com/golang/protobuf/proto"
)

const (
	//message
	PB_BYTES_MESSAGE  = uint8(0x01)
	PB_STRING_MESSAGE = uint8(0x02)
)

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

func MarshalMessage(header *Header, msgType uint8, body interface{}) []byte {
	switch msgType {
	case PB_BYTES_MESSAGE:
		message := &BytesMessage{}
		message.Header = header
		message.Body = body.([]byte)

		data, err := proto.Marshal(message)
		if nil != err {
			log.GetLogger().Errorf("Marshall Bytes Message Error |%s|%d|%s\n", header, msgType, err)
		}
		return data
	case PB_STRING_MESSAGE:
		message := &StringMessage{}
		message.Header = header
		message.Body = proto.String(body.(string))
		data, err := proto.Marshal(message)
		if nil != err {
			log.GetLogger().Errorf("Marshall String Message Error |%s|%d|%s\n", header, msgType, err)
		}
		return data
	}
	return nil
}
