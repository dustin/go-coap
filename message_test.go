package coap

import (
	"bytes"
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

func TestEncodeMessageLargeOptionGap(t *testing.T) {
	req := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
	}

	req.AddOption(ContentFormat, TextPlain)
	req.AddOption(ProxyURI, "u")

	_, err := req.encode()
	if err != ErrOptionGapTooLarge {
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

	data, err := req.encode()
	if err != nil {
		t.Fatalf("Error encoding request: %v", err)
	}

	// Inspected by hand.
	exp := []byte{
		0x40, 0x1, 0x30, 0x39, 0x46, 0x77,
		0x65, 0x65, 0x74, 0x61, 0x67, 0xa1, 0x3,
	}
	if !reflect.DeepEqual(exp, data) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, data)
	}
}

func TestEncodeMessageSmallWithPayload(t *testing.T) {
	req := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
		Payload:   []byte("hi"),
	}

	req.AddOption(ETag, []byte("weetag"))
	req.AddOption(MaxAge, 3)

	data, err := req.encode()
	if err != nil {
		t.Fatalf("Error encoding request: %v", err)
	}

	// Inspected by hand.
	exp := []byte{
		0x40, 0x1, 0x30, 0x39, 0x46, 0x77,
		0x65, 0x65, 0x74, 0x61, 0x67, 0xa1, 0x3,
		0xff, 'h', 'i',
	}
	if !reflect.DeepEqual(exp, data) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, data)
	}
}

func TestDecodeMessageSmallWithPayload(t *testing.T) {
	input := []byte{
		0x40, 0x1, 0x30, 0x39, 0x21, 0x3,
		0x26, 0x77, 0x65, 0x65, 0x74, 0x61, 0x67,
		0xff, 'h', 'i',
	}

	msg, err := parseMessage(input)
	if err != nil {
		t.Fatalf("Error parsing message: %v", err)
	}

	if msg.Type != Confirmable {
		t.Errorf("Expected message type confirmable, got %v", msg.Type)
	}
	if msg.Code != GET {
		t.Errorf("Expected message code GET, got %v", msg.Code)
	}
	if msg.MessageID != 12345 {
		t.Errorf("Expected message ID 12345, got %v", msg.MessageID)
	}

	if !bytes.Equal(msg.Payload, []byte("hi")) {
		t.Errorf("Incorrect payload: %q", msg.Payload)
	}
}

func TestEncodeMessageVerySmall(t *testing.T) {
	req := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
	}
	req.SetPathString("x")

	data, err := req.encode()
	if err != nil {
		t.Fatalf("Error encoding request: %v", err)
	}

	// Inspected by hand.
	exp := []byte{
		0x40, 0x1, 0x30, 0x39, 0xb1, 0x78,
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
		b, err := (*m).encode()
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

	data, err := req.encode()
	if err != nil {
		t.Fatalf("Error encoding request: %v", err)
	}

	// Inspected by hand.
	exp := []byte{
		0x40, 0x1, 0x30, 0x39, 0xbf, 0x17, 0x74, 0x68, 0x69,
		0x73, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x5f, 0x69, 0x73,
		0x5f, 0x6c, 0x6f, 0x6e, 0x67, 0x65, 0x72, 0x5f, 0x74,
		0x68, 0x61, 0x6e, 0x5f, 0x66, 0x69, 0x66, 0x74, 0x65,
		0x65, 0x6e, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x73,
	}
	if !reflect.DeepEqual(exp, data) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, data)
	}
}

func TestDecodeLargePath(t *testing.T) {
	data := []byte{
		0x40, 0x1, 0x30, 0x39, 0xbf, 0x17, 0x74, 0x68,
		0x69, 0x73, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x5f, 0x69, 0x73,
		0x5f, 0x6c, 0x6f, 0x6e, 0x67, 0x65, 0x72, 0x5f, 0x74, 0x68,
		0x61, 0x6e, 0x5f, 0x66, 0x69, 0x66, 0x74, 0x65, 0x65, 0x6e,
		0x5f, 0x62, 0x79, 0x74, 0x65, 0x73,
	}

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
		b, _ := exp.encode()
		t.Fatalf("Expected\n%#v\ngot\n%#v\nfor%#v", exp, req, b)
	}
}

func TestDecodeMessageSmaller(t *testing.T) {
	data := []byte{
		0x40, 0x1, 0x30, 0x39, 0x46, 0x77,
		0x65, 0x65, 0x74, 0x61, 0x67, 0xa1, 0x3,
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

	exp.SetOption(ETag, []byte("weetag"))
	exp.SetOption(MaxAge, uint32(3))

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

/*
    0                   1                   2                   3
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   | 1 | 0 |   0   |     GET=1     |          MID=0x7d34           |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |  11   |  11   |      "temperature" (11 B) ...                 |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
func TestExample1(t *testing.T) {
	input := append([]byte{0x40, 1, 0x7d, 0x34,
		(11 << 4) | 11}, []byte("temperature")...)

	msg, err := parseMessage(input)
	if err != nil {
		t.Fatalf("Error parsing message: %v", err)
	}

	if msg.Type != Confirmable {
		t.Errorf("Expected message type confirmable, got %v", msg.Type)
	}
	if msg.Code != GET {
		t.Errorf("Expected message code GET, got %v", msg.Code)
	}
	if msg.MessageID != 0x7d34 {
		t.Errorf("Expected message ID 0x7d34, got 0x%x", msg.MessageID)
	}

	if msg.Option(URIPath).(string) != "temperature" {
		t.Errorf("Incorrect uri path: %q", msg.Option(URIPath))
	}

	if len(msg.Token) > 0 {
		t.Errorf("Incorrect token: %x", msg.Token)
	}
	if len(msg.Payload) > 0 {
		t.Errorf("Incorrect payload: %q", msg.Payload)
	}
}

/*
    0                   1                   2                   3
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   | 1 | 2 |   0   |    2.05=69    |          MID=0x7d34           |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |1 1 1 1 1 1 1 1|      "22.3 C" (6 B) ...
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
func TestExample1Res(t *testing.T) {
	input := append([]byte{0x60, 69, 0x7d, 0x34, 0xff},
		[]byte("22.3 C")...)

	msg, err := parseMessage(input)
	if err != nil {
		t.Fatalf("Error parsing message: %v", err)
	}

	if msg.Type != Acknowledgement {
		t.Errorf("Expected message type confirmable, got %v", msg.Type)
	}
	if msg.Code != Content {
		t.Errorf("Expected message code Content, got %v", msg.Code)
	}
	if msg.MessageID != 0x7d34 {
		t.Errorf("Expected message ID 0x7d34, got 0x%x", msg.MessageID)
	}

	if len(msg.Token) > 0 {
		t.Errorf("Incorrect token: %x", msg.Token)
	}
	if !bytes.Equal(msg.Payload, []byte("22.3 C")) {
		t.Errorf("Incorrect payload: %q", msg.Payload)
	}
}
