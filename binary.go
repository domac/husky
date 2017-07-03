package husky

import (
	b "encoding/binary"
	"io"
)

var (
	BIG_BYTE_ORDER    = b.BigEndian    //大端序类型
	LITTLE_BYTE_ORDER = b.LittleEndian //小端序类型
)

//字节序-读
func Read(r io.Reader, order b.ByteOrder, data interface{}) error {
	arr, ok := data.([]uint8)
	if ok {
		if _, err := io.ReadFull(r, arr); err != nil {
			return err
		}
		return nil
	} else {
		return b.Read(r, order, data)
	}
}

//字节序-写
func Write(w io.Writer, order b.ByteOrder, data interface{}) error {
	arr, ok := data.([]uint8)
	if ok {
		_, err := w.Write(arr)
		return err
	} else {
		return b.Write(w, order, data)
	}
}
