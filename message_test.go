package coap

import (
	"fmt"
	"reflect"
	"testing"
)

func TestEncodeMessageSmall(t *testing.T) {
	req := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
		Options: Options{
			Option{ETag, []byte("weetag")},
			Option{MaxAge, 3},
		},
	}

	data, err := encodeMessage(req)
	if err != nil {
		t.Fatalf("Error encoding request: %v", err)
	}

	// Inspected by hand.
	exp := []byte{
		0x42, 0x1, 0x30, 0x39, 0x21, 0x3,
		0x26, 0x77, 0x65, 0x65, 0x74, 0x61, 0x67,
	}
	if !reflect.DeepEqual(exp, data) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, data)
	}
}

func TestEncodeLargePath(t *testing.T) {
	req := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
	}
	req.SetPath("this_path_is_longer_than_fifteen_bytes")

	data, err := encodeMessage(req)
	if err != nil {
		t.Fatalf("Error encoding request: %v", err)
	}

	// Inspected by hand.
	exp := []byte{
		0x41, 0x1, 0x30, 0x39, 0x9f, 0x17, 0x74, 0x68,
		0x69, 0x73, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x5f,
		0x69, 0x73, 0x5f, 0x6c, 0x6f, 0x6e, 0x67, 0x65,
		0x72, 0x5f, 0x74, 0x68, 0x61, 0x6e, 0x5f, 0x66,
		0x69, 0x66, 0x74, 0x65, 0x65, 0x6e, 0x5f, 0x62,
		0x79, 0x74, 0x65, 0x73}
	if !reflect.DeepEqual(exp, data) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, data)
	}
}

func TestDecodeLargePath(t *testing.T) {
	data := []byte{
		0x41, 0x1, 0x30, 0x39, 0x9f, 0x17, 0x74, 0x68,
		0x69, 0x73, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x5f,
		0x69, 0x73, 0x5f, 0x6c, 0x6f, 0x6e, 0x67, 0x65,
		0x72, 0x5f, 0x74, 0x68, 0x61, 0x6e, 0x5f, 0x66,
		0x69, 0x66, 0x74, 0x65, 0x65, 0x6e, 0x5f, 0x62,
		0x79, 0x74, 0x65, 0x73}

	req, err := parseMessage(data)
	if err != nil {
		t.Fatalf("Error parsing request: %v", err)
	}

	path := "this_path_is_longer_than_fifteen_bytes"

	exp := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
		Options: Options{
			{URIPath, path},
		},
	}

	if fmt.Sprintf("%#v", exp) != fmt.Sprintf("%#v", req) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, req)
	}
}

func TestDecodeMessageSmaller(t *testing.T) {
	data := []byte{
		0x42, 0x1, 0x30, 0x39, 0x21, 0x3,
		0x26, 0x77, 0x65, 0x65, 0x74, 0x61, 0x67,
	}

	req, err := parseMessage(data)
	if err != nil {
		t.Fatalf("Error parsing request: %v", err)
	}

	exp := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
		Options: Options{
			Option{MaxAge, uint32(3)},
			Option{ETag, []byte("weetag")},
		},
	}

	if fmt.Sprintf("%#v", exp) != fmt.Sprintf("%#v", req) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, req)
	}
}

func TestByteEncoding(t *testing.T) {
	tests := []struct {
		Value    uint32
		Expected []byte
	}{
		{0, []byte{}},
		{13, []byte{13}},
		{1024, []byte{4, 0}},
		{984284, []byte{0x0f, 0x04, 0xdc}},
		{823958824, []byte{0x31, 0x1c, 0x9d, 0x28}},
	}

	for _, v := range tests {
		got := encodeInt(v.Value)
		if !reflect.DeepEqual(got, v.Expected) {
			t.Fatalf("Expected %#v, got %#v for %v",
				v.Expected, got, v.Value)
		}
	}
}

func TestByteDecoding(t *testing.T) {
	tests := []struct {
		Value uint32
		Bytes []byte
	}{
		{0, []byte{}},
		{0, []byte{0}},
		{0, []byte{0, 0}},
		{0, []byte{0, 0, 0}},
		{0, []byte{0, 0, 0, 0}},
		{13, []byte{13}},
		{13, []byte{0, 13}},
		{13, []byte{0, 0, 13}},
		{13, []byte{0, 0, 0, 13}},
		{1024, []byte{4, 0}},
		{1024, []byte{4, 0}},
		{1024, []byte{0, 4, 0}},
		{1024, []byte{0, 0, 4, 0}},
		{984284, []byte{0x0f, 0x04, 0xdc}},
		{984284, []byte{0, 0x0f, 0x04, 0xdc}},
		{823958824, []byte{0x31, 0x1c, 0x9d, 0x28}},
	}

	for _, v := range tests {
		got := decodeInt(v.Bytes)
		if v.Value != got {
			t.Fatalf("Expected %v, got %v for %#v",
				v.Value, got, v.Bytes)
		}
	}
}
