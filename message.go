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

var InvalidTokenLen = errors.New("Invalid token length")
var OptionTooLong = errors.New("Option is too long")
var OptionGapTooLarge = errors.New("Option gap too large")

type OptionID uint8

/*
   +-----+----+---+---+---+----------------+--------+--------+---------+
   | No. | C  | U | N | R | Name           | Format | Length | Default |
   +-----+----+---+---+---+----------------+--------+--------+---------+
   |   1 | x  |   |   | x | If-Match       | opaque | 0-8    | (none)  |
   |   3 | x  | x | - |   | Uri-Host       | string | 1-255  | (see    |
   |     |    |   |   |   |                |        |        | below)  |
   |   4 |    |   |   | x | ETag           | opaque | 1-8    | (none)  |
   |   5 | x  |   |   |   | If-None-Match  | empty  | 0      | (none)  |
   |   7 | x  | x | - |   | Uri-Port       | uint   | 0-2    | (see    |
   |     |    |   |   |   |                |        |        | below)  |
   |   8 |    |   |   | x | Location-Path  | string | 0-255  | (none)  |
   |  11 | x  | x | - | x | Uri-Path       | string | 0-255  | (none)  |
   |  12 |    |   |   |   | Content-Format | uint   | 0-2    | (none)  |
   |  14 |    | x | - |   | Max-Age        | uint   | 0-4    | 60      |
   |  15 | x  | x | - | x | Uri-Query      | string | 0-255  | (none)  |
   |  17 | x  |   |   |   | Accept         | uint   | 0-2    | (none)  |
   |  20 |    |   |   | x | Location-Query | string | 0-255  | (none)  |
   |  35 | x  | x | - |   | Proxy-Uri      | string | 1-1034 | (none)  |
   |  39 | x  | x | - |   | Proxy-Scheme   | string | 1-255  | (none)  |
   |  60 |    |   | x |   | Size1          | uint   | 0-4    | (none)  |
   +-----+----+---+---+---+----------------+--------+--------+---------+
*/

const (
	IfMatch       = OptionID(1)
	URIHost       = OptionID(3)
	ETag          = OptionID(4)
	IfNoneMatch   = OptionID(5)
	URIPort       = OptionID(7)
	LocationPath  = OptionID(8)
	URIPath       = OptionID(11)
	ContentFormat = OptionID(12)
	MaxAge        = OptionID(14)
	URIQuery      = OptionID(15)
	Accept        = OptionID(17)
	LocationQuery = OptionID(20)
	ProxyURI      = OptionID(35)
	ProxyScheme   = OptionID(39)
	Size1         = OptionID(60)
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

	Token, Payload []byte

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

// Get the first value for the given option ID.
func (m Message) Option(o OptionID) interface{} {
	for _, v := range m.opts {
		if o == v.ID {
			return v.Value
		}
	}
	return nil
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
	for s[0] == '/' {
		s = s[1:]
	}
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
	tmpbuf := []byte{0, 0}
	binary.BigEndian.PutUint16(tmpbuf, m.MessageID)

	/*
	     0                   1                   2                   3
	    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	   |Ver| T |  TKL  |      Code     |          Message ID           |
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	   |   Token (if any, TKL bytes) ...
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	   |   Options (if any) ...
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	   |1 1 1 1 1 1 1 1|    Payload (if any) ...
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	*/

    tknlen := uint8(len(m.Token))
    if (tknlen > 8) {
        tknlen = 8
    }

	buf := bytes.Buffer{}
	buf.Write([]byte{
		(1 << 6) | (uint8(m.Type) << 4) | tknlen),
		byte(m.Code),
		tmpbuf[0], tmpbuf[1],
	})

    
	buf.Write(m.Token[:tknlen])

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
        optdelta := o.ID - prev
		optlen := len(b)

        var optlenbytes byte[]
        if optlen >= 269 {
            optlenbytes = encodeInt(optlen - 269);
            optlen = 14;
        } else if optlen >= 13 {
            optlenbytes = encodeInt(optlen - 13);
            optlen = 13;
        } else {
            optlenbytes = nil;
        }
        
        var optdeltabytes byte[]
        if optlen >= 269 {
            optdeltabytes = encodeInt(optdelta - 269);
            optdelta = 14;
        } else if optlen >= 13 {
            optdeltabytes = encodeInt(optdelta - 13);
            optdelta = 13;
        } else {
            optdeltabytes = nil;
        }

        optdeltalenbyte := byte((optdelta << 4) + optlen);

        buf.Write(byte[](optdeltalenbyte))
        buf.Write(optdeltabytes)
        buf.Write(optlenbytes)
		buf.Write(b)
		prev = int(o.ID)
	}

	if len(m.Payload) > 0 {
		buf.Write([]byte{0xff})
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
	tokenLen := uint8(data[0] & 0xf)
	if tokenLen > 8 {
		return rv, InvalidTokenLen
	}

	rv.Code = COAPCode(data[1])
	rv.MessageID = binary.BigEndian.Uint16(data[2:4])

	b := data[4:]
    
    // Token
    rv.Token = b[:tokenLen]
    b = b[tokenLen:]

    // Options
	prev := 0
	for len(b) > 0 {
        optlen := (b[0] >> 4)
        optdelta := b[0] & 0xf
        b = b[1:]

        if (optlen == 15) || (optdelta == 15) {
            if (optlen == 15) && (optdelta == 15) {
                break;
            }
            else return rv, errors.New("Invalid Option: Len xor Delta was 15")
        }

        if optdelta == 14 {
            optdelta = decodeInt(b[:2]) - 269;
            b = b[2:]
        } else if optdelta == 13 {
            optdelta = decodeInt(b[:1]) - 13;
            b = b[1:]
        }

        if optlen == 14 {
            optlen = decodeInt(b[:2]) - 269;
            b = b[2:]
        } else if optlen == 13 {
            optlen = decodeInt(b[:1]) - 13;
            b = b[1:]
        }

        if len(b) < optlen {
            return rv, errors.New("Truncated option")
        }
        
        var opval interface{} = b[:optlen]

		oid := OptionID(prev + optdelta))
		switch oid {
		case URIPort, ContentFormat, MaxAge, Accept, Size1:
			opval = decodeInt(b[:optlen])
		case URIHost, LocationPath, URIPath, URIQuery, LocationQuery,
			ProxyURI, ProxyScheme:
			opval = string(b[:optlen])
		}

		option := option{
			ID:    oid,
			Value: opval,
		}
		b = b[optlen:]
		prev = int(oid)

		rv.opts = append(rv.opts, option)
	}

	rv.Payload = b
	return rv, nil
}
