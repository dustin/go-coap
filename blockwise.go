package coap

import (
	"bytes"
	"encoding/binary"
	"math"
)

// Block represents a block in a block-wise transfer
type Block struct {
	more bool
	num  uint32
	size uint32
}

// MarshalBinary produces the binary form of this Block
func (b *Block) MarshalBinary() uint32 {
	value := b.num << 4
	value |= uint32((math.Log(float64(b.size)) / math.Log(2)) - 4)

	if b.more {
		value |= 0x8
	}

	buf := bytes.Buffer{}
	binary.Write(&buf, binary.BigEndian, value)

	return binary.BigEndian.Uint32(buf.Bytes())
}

// ParseBlock parses the given binary data as a Block
func ParseBlock(data []interface{}) *Block {
	if len(data) != 1 {
		return nil
	}

	b := &Block{
		more: false,
		num:  data[0].(uint32) >> 4,
		size: uint32(math.Pow(2, float64((data[0].(uint32)&0x07)+4))),
	}

	if data[0].(uint32)&(0x01<<3) > 0 {
		b.more = true
	}

	return b
}
