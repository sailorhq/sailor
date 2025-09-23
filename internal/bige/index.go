package bige

import (
	"bytes"
	"encoding/binary"
)

func ByteFromUInt32(num uint32) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, &num)
	return buf.Bytes()
}

func UInt32FromByte(data []byte) uint32 {
	var num uint32
	binary.Read(bytes.NewBuffer(data), binary.BigEndian, &num)
	return num
}
