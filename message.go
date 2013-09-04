package coap

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type COAPType uint8

const (
	Confirmable     = COAPType(0)
	NonConfirmable  = COAPType(1)
	Acknowledgement = COAPType(2)
	Reset           = COAPType(3)
)

var typeNames = [256]string{
	Confirmable:     "Confirmable",
	NonConfirmable:  "NonConfirmable",
	Acknowledgement: "Acknowledgement",
	Reset:           "Reset",
}

func init() {
	for i := range typeNames {
		if typeNames[i] == "" {
			typeNames[i] = fmt.Sprintf("Unknown (0x%x)", i)
		}
	}
}

func (t COAPType) String() string {
	return typeNames[t]
}

type COAPCode uint8

// Request Codes
const (
	GET       COAPCode = 1
	POST      COAPCode = 2
	PUT       COAPCode = 3
	DELETE    COAPCode = 4
	SUBSCRIBE COAPCode = 5
)

// Response Codes
const (
	Created               COAPCode = 65
	Deleted               COAPCode = 66
	Valid                 COAPCode = 67
	Changed               COAPCode = 68
	Content               COAPCode = 69
	BadRequest            COAPCode = 128
	Unauthorized          COAPCode = 129
	BadOption             COAPCode = 130
	Forbidden             COAPCode = 131
	NotFound              COAPCode = 132
	MethodNotAllowed      COAPCode = 133
	NotAcceptable         COAPCode = 134
	PreconditionFailed    COAPCode = 140
	RequestEntityTooLarge COAPCode = 141
	UnsupportedMediaType  COAPCode = 143
	InternalServerError   COAPCode = 160
	NotImplemented        COAPCode = 161
	BadGateway            COAPCode = 162
	ServiceUnavailable    COAPCode = 163
	GatewayTimeout        COAPCode = 164
	ProxyingNotSupported  COAPCode = 165
)

var codeNames = [256]string{
	GET:                   "GET",
	POST:                  "POST",
	PUT:                   "PUT",
	DELETE:                "DELETE",
	SUBSCRIBE:             "SUBSCRIBE",
	Created:               "Created",
	Deleted:               "Deleted",
	Valid:                 "Valid",
	Changed:               "Changed",
	Content:               "Content",
	BadRequest:            "BadRequest",
	Unauthorized:          "Unauthorized",
	BadOption:             "BadOption",
	Forbidden:             "Forbidden",
	NotFound:              "NotFound",
	MethodNotAllowed:      "MethodNotAllowed",
	NotAcceptable:         "NotAcceptable",
	PreconditionFailed:    "PreconditionFailed",
	RequestEntityTooLarge: "RequestEntityTooLarge",
	UnsupportedMediaType:  "UnsupportedMediaType",
	InternalServerError:   "InternalServerError",
	NotImplemented:        "NotImplemented",
	BadGateway:            "BadGateway",
	ServiceUnavailable:    "ServiceUnavailable",
	GatewayTimeout:        "GatewayTimeout",
	ProxyingNotSupported:  "ProxyingNotSupported",
}

func init() {
	for i := range codeNames {
		if codeNames[i] == "" {
			codeNames[i] = fmt.Sprintf("Unknown (0x%x)", i)
		}
	}
}

func (c COAPCode) String() string {
	return codeNames[c]
}

var TooManyoptions = errors.New("Too many options")
var OptionTooLong = errors.New("Option is too long")
var OptionGapTooLarge = errors.New("Option gap too large")

type OptionID uint8

const (
	ContentType   = OptionID(1)
	MaxAge        = OptionID(2)
	ProxyURI      = OptionID(3)
	ETag          = OptionID(4)
	URIHost       = OptionID(5)
	LocationPath  = OptionID(6)
	URIPort       = OptionID(7)
	LocationQuery = OptionID(8)
	URIPath       = OptionID(9)
	Token         = OptionID(11)
	Accept        = OptionID(12)
	IfMatch       = OptionID(13)
	UriQuery      = OptionID(15)
	IfNoneMatch   = OptionID(21)
)

type MediaType byte

const (
	TextPlain     = MediaType(0)  // text/plain;charset=utf-8
	AppLinkFormat = MediaType(40) // application/link-format
	AppXML        = MediaType(41) // application/xml
	AppOctets     = MediaType(42) // application/octet-stream
	AppExi        = MediaType(47) // application/exi
	AppJSON       = MediaType(50) // application/json
)

/*
   +-----+---+---+----------------+--------+---------+-------------+
   | No. | C | R | Name           | Format | Length  | Default     |
   +-----+---+---+----------------+--------+---------+-------------+
   |   1 | x |   | Content-Type   | uint   | 0-2 B   | (none)      |
   |   2 |   |   | Max-Age        | uint   | 0-4 B   | 60          |
   |   3 | x | x | Proxy-Uri      | string | 1-270 B | (none)      |
   |   4 |   | x | ETag           | opaque | 1-8 B   | (none)      |
   |   5 | x |   | Uri-Host       | string | 1-270 B | (see below) |
   |   6 |   | x | Location-Path  | string | 0-270 B | (none)      |
   |   7 | x |   | Uri-Port       | uint   | 0-2 B   | (see below) |
   |   8 |   | x | Location-Query | string | 0-270 B | (none)      |
   |   9 | x | x | Uri-Path       | string | 0-270 B | (none)      |
   |  11 | x |   | Token          | opaque | 1-8 B   | (empty)     |
   |  12 |   | x | Accept         | uint   | 0-2 B   | (none)      |
   |  13 | x | x | If-Match       | opaque | 0-8 B   | (none)      |
   |  15 | x | x | Uri-Query      | string | 0-270 B | (none)      |
   |  21 | x |   | If-None-Match  | empty  | 0 B     | (none)      |
   +-----+---+---+----------------+--------+---------+-------------+
*/

type option struct {
	ID    OptionID
	Value interface{}
}

func encodeInt(v uint32) []byte {
	switch {
	case v == 0:
		return nil
	case v < 256:
		return []byte{byte(v)}
	case v < 65536:
		rv := []byte{0, 0}
		binary.BigEndian.PutUint16(rv, uint16(v))
		return rv
	case v < 16777216:
		rv := []byte{0, 0, 0, 0}
		binary.BigEndian.PutUint32(rv, uint32(v))
		return rv[1:]
	default:
		rv := []byte{0, 0, 0, 0}
		binary.BigEndian.PutUint32(rv, uint32(v))
		return rv
	}
}

func decodeInt(b []byte) uint32 {
	tmp := []byte{0, 0, 0, 0}
	copy(tmp[4-len(b):], b)
	return binary.BigEndian.Uint32(tmp)
}

func (o option) toBytes() []byte {
	var v uint32

	switch i := o.Value.(type) {
	case string:
		return []byte(i)
	case []byte:
		return i
	case MediaType:
		v = uint32(i)
	case int:
		v = uint32(i)
	case int32:
		v = uint32(i)
	case uint:
		v = uint32(i)
	case uint32:
		v = i
	default:
		panic(fmt.Errorf("Invalid type for option %x: %T (%v)",
			o.ID, o.Value, o.Value))
	}

	return encodeInt(v)
}

type options []option

func (o options) Len() int {
	return len(o)
}

func (o options) Less(i, j int) bool {
	if o[i].ID == o[j].ID {
		return i < j
	}
	return o[i].ID < o[j].ID
}

func (o options) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

func (o options) Minus(oid OptionID) options {
	rv := options{}
	for _, opt := range o {
		if opt.ID != oid {
			rv = append(rv, opt)
		}
	}
	return rv
}

// A CoAP message.
type Message struct {
	Type      COAPType
	Code      COAPCode
	MessageID uint16

	Payload []byte

	opts options
}

// Return True if this message is confirmable.
func (m Message) IsConfirmable() bool {
	return m.Type == Confirmable
}

// Get all the values for the given option.
func (m Message) Options(o OptionID) []interface{} {
	var rv []interface{}

	for _, v := range m.opts {
		if o == v.ID {
			rv = append(rv, v.Value)
		}
	}

	return rv
}

func (m Message) optionStrings(o OptionID) []string {
	var rv []string
	for _, o := range m.Options(o) {
		rv = append(rv, o.(string))
	}
	return rv
}

// Get the Path set on this message if any.
func (m Message) Path() []string {
	return m.optionStrings(URIPath)
}

// Get a path as a / separated string.
func (m Message) PathString() string {
	return strings.Join(m.Path(), "/")
}

// Set a path by a / separated string.
func (m *Message) SetPathString(s string) {
	m.SetPath(strings.Split(s, "/"))
}

// Update or add a LocationPath attribute on this message.
func (m *Message) SetPath(s []string) {
	m.RemoveOption(URIPath)
	for _, p := range s {
		m.AddOption(URIPath, p)
	}
}

// Remove all references to an option
func (m *Message) RemoveOption(opId OptionID) {
	m.opts = m.opts.Minus(opId)
}

// Add an option.
func (m *Message) AddOption(opId OptionID, val interface{}) {
	m.opts = append(m.opts, option{opId, val})
}

// Set an option, discarding any previous value
func (m *Message) SetOption(opId OptionID, val interface{}) {
	m.RemoveOption(opId)
	m.AddOption(opId, val)
}

func (m *Message) encode() ([]byte, error) {
	if len(m.opts) > 14 {
		return nil, TooManyoptions
	}

	tmpbuf := []byte{0, 0}
	binary.BigEndian.PutUint16(tmpbuf, m.MessageID)

	/*
	     0                   1                   2                   3
	    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	   |Ver| T |  OC   |      Code     |          Message ID           |
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	   |   options (if any) ...
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	   |   Payload (if any) ...
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	*/

	buf := bytes.Buffer{}
	buf.Write([]byte{
		(1 << 6) | (uint8(m.Type) << 4) | uint8(0xf&len(m.opts)),
		byte(m.Code),
		tmpbuf[0], tmpbuf[1],
	})

	/*
	     0   1   2   3   4   5   6   7
	   +---+---+---+---+---+---+---+---+
	   | Option Delta  |    Length     | for 0..14
	   +---+---+---+---+---+---+---+---+
	   |   Option Value ...
	   +---+---+---+---+---+---+---+---+
	                                               for 15..270:
	   +---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+
	   | Option Delta  | 1   1   1   1 |          Length - 15          |
	   +---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+
	   |   Option Value ...
	   +---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+
	*/

	sort.Sort(&m.opts)

	prev := 0
	for _, o := range m.opts {
		b := o.toBytes()
		if len(b) > 15 {
			buf.Write([]byte{
				byte(int(o.ID)-prev)<<4 | 15,
				byte(len(b) - 15),
			})
		} else {
			buf.Write([]byte{byte(int(o.ID)-prev)<<4 | byte(len(b))})
		}
		if int(o.ID)-prev > 15 {
			return nil, OptionGapTooLarge
		}

		buf.Write(b)
		prev = int(o.ID)
	}

	buf.Write(m.Payload)

	return buf.Bytes(), nil
}

func parseMessage(data []byte) (rv Message, err error) {
	if len(data) < 6 {
		return rv, errors.New("Short packet")
	}

	if data[0]>>6 != 1 {
		return rv, errors.New("Invalid version")
	}

	rv.Type = COAPType((data[0] >> 4) & 0x3)
	opCount := int(data[0] & 0xf)
	if opCount > 14 {
		return rv, TooManyoptions
	}

	rv.Code = COAPCode(data[1])
	rv.MessageID = binary.BigEndian.Uint16(data[2:4])

	b := data[4:]
	prev := 0
	for i := 0; i < opCount && len(b) > 0; i++ {
		oid := OptionID(prev + int(b[0]>>4))
		l := int(b[0] & 0xf)
		b = b[1:]
		if l > 14 {
			l += int(b[0])
			b = b[1:]
		}
		if len(b) < l {
			return rv, errors.New("Truncated")
		}
		var opval interface{} = b[:l]
		switch oid {
		case ContentType,
			MaxAge,
			URIPort,
			Accept:
			opval = decodeInt(b[:l])
		case ProxyURI, URIHost, LocationPath, LocationQuery, URIPath, UriQuery:
			opval = string(b[:l])
		}

		option := option{
			ID:    oid,
			Value: opval,
		}
		b = b[l:]
		prev = int(option.ID)

		rv.opts = append(rv.opts, option)
	}

	rv.Payload = b
	return rv, nil
}
