package coap

import (
	"bytes"
	"encoding/binary"
	"math"
)

// Block represents a block in a block-wise transfer
type Block struct {
	// More flag
	More bool
	// Block number
	Num uint32
	// Block size
	Size uint32
}

// MarshalBinary produces the binary form of this Block
func (b *Block) MarshalBinary() uint32 {
	value := b.Num << 4
	value |= uint32((math.Log(float64(b.Size)) / math.Log(2)) - 4)

	if b.More {
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
		More: false,
		Num:  data[0].(uint32) >> 4,
		Size: uint32(math.Pow(2, float64((data[0].(uint32)&0x07)+4))),
	}

	if data[0].(uint32)&(0x01<<3) > 0 {
		b.More = true
	}

	return b
}
