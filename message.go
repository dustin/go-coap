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

const (
	GET       = 1
	POST      = 2
	PUT       = 3
	DELETE    = 4
	SUBSCRIBE = 5
)

const (
	Created               = 65
	Deleted               = 66
	Valid                 = 67
	Changed               = 68
	Content               = 69
	BadRequest            = 128
	Unauthorized          = 129
	BadOption             = 130
	Forbidden             = 131
	NotFound              = 132
	MethodNotAllowed      = 133
	NotAcceptable         = 134
	PreconditionFailed    = 140
	RequestEntityTooLarge = 141
	UnsupportedMediaType  = 143
	InternalServerError   = 160
	NotImplemented        = 161
	BadGateway            = 162
	ServiceUnavailable    = 163
	GatewayTimeout        = 164
	ProxyingNotSupported  = 165
)

var TooManyOptions = errors.New("Too many options")
var OptionTooLong = errors.New("Option is too long")

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

type Option struct {
	ID    OptionID
	Value interface{}
}

func encodeInt(v uint32) []byte {
	switch {
	case v == 0:
		return []byte{}
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

func (o Option) toBytes() []byte {
	var v uint32

	switch i := o.Value.(type) {
	case string:
		return []byte(i)
	case []byte:
		return i
	case int:
		v = uint32(i)
	case int32:
		v = uint32(i)
	case uint:
		v = uint32(i)
	case uint32:
		v = i
	default:
		panic(fmt.Errorf("Invalid type for option %x", o.ID))
	}

	return encodeInt(v)
}

type Options []Option

func (o Options) Len() int {
	return len(o)
}

func (o Options) Less(i, j int) bool {
	if o[i].ID == o[j].ID {
		return i < j
	}
	return o[i].ID < o[j].ID
}

func (o Options) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

func (o Options) Minus(oid OptionID) Options {
	rv := Options{}
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
	Code      uint8
	MessageID uint16

	Options Options

	Payload []byte
}

// Return True if this message is confirmable.
func (m Message) IsConfirmable() bool {
	return m.Type == Confirmable
}

// Get the Path set on this message if any.
func (m Message) Path() []string {
	rv := []string{}
	for _, o := range m.Options {
		if o.ID == URIPath {
			rv = append(rv, o.Value.(string))
		}
	}
	return rv
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
	m.Options = m.Options.Minus(URIPath)
	for _, p := range s {
		m.Options = append(m.Options, Option{URIPath, p})
	}
}

func encodeMessage(r Message) ([]byte, error) {
	if len(r.Options) > 14 {
		return []byte{}, TooManyOptions
	}

	tmpbuf := []byte{0, 0}
	binary.BigEndian.PutUint16(tmpbuf, r.MessageID)

	/*
	     0                   1                   2                   3
	    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	   |Ver| T |  OC   |      Code     |          Message ID           |
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	   |   Options (if any) ...
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	   |   Payload (if any) ...
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	*/

	buf := bytes.Buffer{}
	buf.Write([]byte{
		(1 << 6) | (uint8(r.Type) << 4) | uint8(0xf&len(r.Options)),
		byte(r.Code),
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

	sort.Sort(&r.Options)

	prev := 0
	for _, o := range r.Options {
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
			return []byte{}, errors.New("Gap too large")
		}

		buf.Write(b)
		prev = int(o.ID)
	}

	buf.Write(r.Payload)

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
		return rv, TooManyOptions
	}

	rv.Code = data[1]
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

		option := Option{
			ID:    oid,
			Value: opval,
		}
		b = b[l:]
		prev = int(option.ID)

		rv.Options = append(rv.Options, option)
	}

	rv.Payload = b
	return rv, nil
}
