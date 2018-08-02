package coap

import (
	"reflect"
	"testing"
)

func TestBlockMarshalAndUnmarshal(t *testing.T) {
	block := &Block{
		num:  4096 * 1024,
		more: true,
		size: 1024,
	}

	b := ParseBlock([]interface{}{block.MarshalBinary()})

	if !reflect.DeepEqual(*block, *b) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", block, b)
	}
}
