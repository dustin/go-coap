package coap

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net/url"
	"sort"
)

type COAPType uint8

const (
	Confirmable     = COAPType(0)
	NonConfirmable  = COAPType(1)
	Acknowledgement = COAPType(2)
	Reset           = COAPType(3)
)

type COAPMethod uint8

const (
	GET    = COAPMethod(1)
	POST   = COAPMethod(2)
	PUT    = COAPMethod(3)
	DELETE = COAPMethod(4)
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
	UriPath       = OptionID(9)
	Token         = OptionID(11)
	Accept        = OptionID(12)
	IfMatch       = OptionID(13)
	UriQuery      = OptionID(15)
	IfNoneMatch   = OptionID(21)
)

type MediaType uint8

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
	Value []byte
}

type Options []Option

func (o Options) Len() int {
	return len(o)
}

func (o Options) Less(i, j int) bool {
	return o[i].ID < o[j].ID
}

func (o Options) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

type Message struct {
	URL url.URL

	Type      COAPType
	Code      COAPMethod
	MessageID uint16

	Options Options

	Payload []byte
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
		if len(o.Value) > 15 {
			return []byte{}, OptionTooLong
		}
		if int(o.ID)-prev > 15 {
			return []byte{}, errors.New("Gap too large")
		}

		buf.Write([]byte{byte(int(o.ID)-prev)<<4 | byte(len(o.Value))})
		buf.Write(o.Value)
		prev = int(o.ID)
	}

	buf.Write(r.Payload)

	return buf.Bytes(), nil
}

func parseMessage(data []byte) (rv Message, err error) {
	if len(data) < 8 {
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

	rv.Code = COAPMethod(data[1])
	rv.MessageID = binary.BigEndian.Uint16(data[2:4])

	b := data[4:]
	prev := 0
	for i := 0; i < opCount && len(b) > 0; i++ {
		l := int(b[0] & 0xf)
		if l > 14 {
			return rv, OptionTooLong
		}
		if len(b) < l {
			return rv, errors.New("Truncated")
		}
		option := Option{
			ID:    OptionID(prev + int(b[0]>>4)),
			Value: b[1 : l+1],
		}
		b = b[l+1:]
		prev = int(option.ID)

		rv.Options = append(rv.Options, option)
	}

	rv.Payload = b
	return rv, nil
}
