package coap

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"
)

func TestEncodeRequestSmall(t *testing.T) {
	u, err := url.Parse("coap://localhost/")
	if err != nil {
		t.Fatalf("Error parsing URL: %v", err)
	}

	req := Req{
		URL:       *u,
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
		Options: Options{
			Option{ETag, []byte("weetag")},
			Option{MaxAge, []byte{0, 0, 0, 3}},
		},
	}

	data, err := encodeRequest(req)
	if err != nil {
		t.Fatalf("Error encoding request: %v", err)
	}

	// Inspected by hand.
	exp := []byte{
		0x42, 0x1, 0x30, 0x39, 0x24, 0x0, 0x0, 0x0, 0x3,
		0x26, 0x77, 0x65, 0x65, 0x74, 0x61, 0x67,
	}
	if !reflect.DeepEqual(exp, data) {
		t.Fatalf("Expected %#v, got %#v", exp, data)
	}
}

func TestDecodeRequestSmall(t *testing.T) {
	data := []byte{
		0x42, 0x1, 0x30, 0x39, 0x24, 0x0, 0x0, 0x0, 0x3,
		0x26, 0x77, 0x65, 0x65, 0x74, 0x61, 0x67,
	}

	req, err := parseRequest(data)
	if err != nil {
		t.Fatalf("Error parsing request: %v", err)
	}

	exp := Req{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
		Options: Options{
			Option{MaxAge, []byte{0, 0, 0, 3}},
			Option{ETag, []byte("weetag")},
		},
	}

	if fmt.Sprintf("%#v", exp) != fmt.Sprintf("%#v", req) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, req)
	}
}
