package coap

import (
	"fmt"
	"reflect"
	"testing"
)

func TestTypeString(t *testing.T) {
	tests := map[COAPType]string{
		Confirmable:    "Confirmable",
		NonConfirmable: "NonConfirmable",
		255:            "Unknown (0xff)",
	}

	for code, exp := range tests {
		if code.String() != exp {
			t.Errorf("Error on %d, got %v, expected %v",
				code, code, exp)
		}
	}
}

func TestCodeString(t *testing.T) {
	tests := map[COAPCode]string{
		0:             "Unknown (0x0)",
		GET:           "GET",
		POST:          "POST",
		NotAcceptable: "NotAcceptable",
		255:           "Unknown (0xff)",
	}

	for code, exp := range tests {
		if code.String() != exp {
			t.Errorf("Error on %d, got %v, expected %v",
				code, code, exp)
		}
	}
}

func TestEncodeMessageTooManyoptions(t *testing.T) {
	req := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
	}

	options := []option{
		option{ETag, []byte("weetag")},
		option{MaxAge, 3},
		option{ContentType, TextPlain},
		option{IfMatch, "a"},
		option{IfMatch, "b"},
		option{IfMatch, "c"},
		option{IfMatch, "d"},
		option{IfMatch, "e"},
		option{IfMatch, "f"},
		option{IfNoneMatch, "z"},
		option{IfNoneMatch, "y"},
		option{IfNoneMatch, "x"},
		option{IfNoneMatch, "w"},
		option{IfNoneMatch, "v"},
		option{IfNoneMatch, "u"},
	}

	for _, o := range options {
		req.AddOption(o.ID, o.Value)
	}

	req.SetPathString("/a/b/c/d/e/f/g/h")

	_, err := encodeMessage(req)
	if err != TooManyoptions {
		t.Fatalf("Expected 'too many options', got: %v", err)
	}
}

func TestEncodeMessageLargeOptionGap(t *testing.T) {
	req := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
	}

	req.AddOption(ContentType, TextPlain)
	req.AddOption(IfNoneMatch, "u")

	_, err := encodeMessage(req)
	if err != OptionGapTooLarge {
		t.Fatalf("Expected 'option gap too large', got: %v", err)
	}
}

func TestEncodeMessageSmall(t *testing.T) {
	req := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
	}

	req.AddOption(ETag, []byte("weetag"))
	req.AddOption(MaxAge, 3)

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

func TestEncodeMessageVerySmall(t *testing.T) {
	req := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
	}
	req.SetPathString("x")

	data, err := encodeMessage(req)
	if err != nil {
		t.Fatalf("Error encoding request: %v", err)
	}

	// Inspected by hand.
	exp := []byte{
		0x41, 0x1, 0x30, 0x39, 0x91, 0x78,
	}
	if !reflect.DeepEqual(exp, data) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, data)
	}
}

func TestEncodeSeveral(t *testing.T) {
	tests := map[string][]string{
		"a":   []string{"a"},
		"axe": []string{"axe"},
		"a/b/c/d/e/f/h/g/i/j": []string{"a", "b", "c", "d", "e",
			"f", "h", "g", "i", "j"},
	}
	for p, a := range tests {
		m := &Message{Type: Confirmable, Code: GET, MessageID: 12345}
		m.SetPathString(p)
		b, err := encodeMessage(*m)
		if err != nil {
			t.Errorf("Error encoding %#v", p)
			t.Fail()
			continue
		}
		m2, err := parseMessage(b)
		if err != nil {
			t.Fatalf("Can't parse my own message at %#v: %v", p, err)
		}

		if !reflect.DeepEqual(m2.Path(), a) {
			t.Errorf("Expected %#v, got %#v", a, m2.Path())
			t.Fail()
		}
	}
}

func TestEncodeLargePath(t *testing.T) {
	req := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
	}
	req.SetPathString("this_path_is_longer_than_fifteen_bytes")

	if req.PathString() != "this_path_is_longer_than_fifteen_bytes" {
		t.Fatalf("Didn't get back the same path I posted: %v",
			req.PathString())
	}

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
	}

	exp.SetOption(URIPath, path)

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
	}

	exp.SetOption(MaxAge, uint32(3))
	exp.SetOption(ETag, []byte("weetag"))

	if fmt.Sprintf("%#v", exp) != fmt.Sprintf("%#v", req) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, req)
	}
}

func TestByteEncoding(t *testing.T) {
	tests := []struct {
		Value    uint32
		Expected []byte
	}{
		{0, nil},
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
		{0, nil},
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
