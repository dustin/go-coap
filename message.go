package coap

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// COAPType represents the message type.
type COAPType uint8

const (
	// Confirmable messages require acknowledgements.
	Confirmable = COAPType(0)
	// NonConfirmable messages do not require acknowledgements.
	NonConfirmable = COAPType(1)
	// Acknowledgement is a message type indicating a response to
	// a confirmable message.
	Acknowledgement = COAPType(2)
	// Reset indicates a permanent negative acknowledgement.
	Reset = COAPType(3)
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
			typeNames[i] = fmt.Sprintf("unknown (0x%x)", i)
		}
	}
}

func (t COAPType) String() string {
	return typeNames[t]
}

// COAPCode is the type used for both request and response codes.
type COAPCode uint8

// For RFC1982 (Serial Arithmetic) calculations in 32 bits.
const year68 = 1 << 31

// Request Codes
const (
	GET    COAPCode = 1
	POST   COAPCode = 2
	PUT    COAPCode = 3
	DELETE COAPCode = 4
	// deprecated
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
			codeNames[i] = fmt.Sprintf("unknown (0x%x)", i)
		}
	}
}

func (c COAPCode) String() string {
	return codeNames[c]
}

// Message encoding errors.
var (
	ErrInvalidTokenLen   = errors.New("invalid token length")
	ErrOptionTooLong     = errors.New("option is too long")
	ErrOptionGapTooLarge = errors.New("option gap too large")
)

/*
   +-----+----+---+---+---+----------------+--------+--------+---------+
   | No. | C  | U | N | R | Name           | Format | Length | Default |
   +-----+----+---+---+---+----------------+--------+--------+---------+
   |   1 | x  |   |   | x | If-Match       | opaque | 0-8    | (none)  |
   |   3 | x  | x | - |   | Uri-Host       | string | 1-255  | (see    |
   |     |    |   |   |   |                |        |        | below)  |
   |   4 |    |   |   | x | ETag           | opaque | 1-8    | (none)  |
   |   5 | x  |   |   |   | If-None-Match  | empty  | 0      | (none)  |
   |   6 |    | x | - |   | Observe        | empty/ | 0/0-3  | (none)  |
   |     |    |   |   |   |                | uint   |        |         |
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

// Option IDs.
type OptionID uint8

const (
	IfMatch       = OptionID(1)
	URIHost       = OptionID(3)
	ETag          = OptionID(4)
	IfNoneMatch   = OptionID(5)
	Observe       = OptionID(6)
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

// MediaType specifies the content type of a message.
type MediaType byte

// Content types.
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
	case []string:
		return []byte(strings.Join(i, ""))
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
		debugMsg("invalid type for option %x: %T (%v)",
			o.ID, o.Value, o.Value)
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

// Message is a CoAP message.
type Message struct {
	Type      COAPType
	Code      COAPCode
	MessageID uint16

	Token, Payload []byte

	opts options
}

// TimeToString translates the RRSIG's incep. and expir. times to the
// string representation used when printing the record.
// It takes serial arithmetic (RFC 1982) into observe option.
func TimeToString(t uint32) string {
	mod := ((int64(t) - time.Now().Unix()) / year68) - 1
	if mod < 0 {
		mod = 0
	}
	ti := time.Unix(int64(t)-(mod*year68), 0).UTC()
	return ti.Format("20060102150405")
}

// StringToTime translates the RRSIG's incep. and expir. times from
// string values like "20110403154150" to an 32 bit integer.
// It takes serial arithmetic (RFC 1982) into observe option.
func StringToTime(s string) (uint32, error) {
	t, e := time.Parse("20060102150405", s)
	if e != nil {
		return 0, e
	}
	mod := (t.Unix() / year68) - 1
	if mod < 0 {
		mod = 0
	}
	return uint32(t.Unix() - (mod * year68)), nil
}

// IsConfirmable returns true if this message is confirmable.
func (m Message) IsConfirmable() bool {
	return m.Type == Confirmable
}

// IsObserver returns true if this message is for observe.
func (m Message) IsObserver() bool {
	for _, o := range m.opts {
		return o.ID == Observe
	}
	return false
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

// Option gets the first value for the given option ID.
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

// Path gets the Path set on this message if any.
func (m Message) Path() []string {
	return m.optionStrings(URIPath)
}

// PathString gets a path as a / separated string.
func (m Message) PathString() string {
	return strings.Join(m.Path(), "/")
}

// SetPathString sets a path by a / separated string.
func (m *Message) SetPathString(s string) {
	for s[0] == '/' {
		s = s[1:]
	}
	m.SetPath(strings.Split(s, "/"))
}

// SetPath updates or adds a LocationPath attribute on this message.
func (m *Message) SetPath(s []string) {
	m.RemoveOption(URIPath)
	for _, p := range s {
		m.AddOption(URIPath, p)
	}
}

// RemoveOption removes all references to an option
func (m *Message) RemoveOption(opId OptionID) {
	m.opts = m.opts.Minus(opId)
}

// AddOption adds an option.
func (m *Message) AddOption(opId OptionID, val interface{}) {
	m.opts = append(m.opts, option{opId, val})
}

// SetOption sets an option, discarding any previous value
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

	buf := bytes.Buffer{}
	buf.Write([]byte{
		(1 << 6) | (uint8(m.Type) << 4) | uint8(0xf&len(m.Token)),
		byte(m.Code),
		tmpbuf[0], tmpbuf[1],
	})
	buf.Write(m.Token)

	/*
	     0   1   2   3   4   5   6   7
	   +---------------+---------------+
	   |               |               |
	   |  Option Delta | Option Length |   1 byte
	   |               |               |
	   +---------------+---------------+
	   \                               \
	   /         Option Delta          /   0-2 bytes
	   \          (extended)           \
	   +-------------------------------+
	   \                               \
	   /         Option Length         /   0-2 bytes
	   \          (extended)           \
	   +-------------------------------+
	   \                               \
	   /                               /
	   \                               \
	   /         Option Value          /   0 or more bytes
	   \                               \
	   /                               /
	   \                               \
	   +-------------------------------+

	  http://tools.ietf.org/html/draft-ietf-core-coap-18#section-3.1
	*/
	sort.Sort(&m.opts)

	prev := 0
	for _, o := range m.opts {
		b := o.toBytes()

		optDelta := int(o.ID) - prev
		delta := optionNibble(optDelta)
		l := optionNibble(len(b))
		buf.Write([]byte{0xFF & (byte(delta)<<4 | byte(l))})
		if delta == 13 {
			buf.Write([]byte{byte(optDelta - 13)})
		} else if delta == 14 {
			buf.Write([]byte{byte(uint8(optDelta-269) >> 8)})
			buf.Write([]byte{byte(0xFF & (optDelta - 269))})
		}
		if l == 13 {
			buf.Write([]byte{byte(len(b) - 13)})
		} else if l == 14 {
			buf.Write([]byte{byte(uint8(len(b)) >> 8)})
			buf.Write([]byte{byte(0xFF & (len(b) - 269))})
		}

		if optDelta > 15 {
			return nil, ErrOptionGapTooLarge
		}
		buf.Write(b)
		prev = int(o.ID)
	}

	if len(m.Payload) > 0 {
		buf.Write([]byte{0xff})
	}

	buf.Write(bytes.Trim(m.Payload, " \x00"))

	return buf.Bytes(), nil
}

func optionNibble(value int) int {
	if value < 13 {
		return (0xFF & value)
	} else if value <= 0xFF+13 {
		return 13
	} else if value <= 0xFFFF+269 {
		return 14
	} else {
		panic("nibble value larger that 526345 is not supported")
	}
}

func parseMessage(data []byte) (rv Message, err error) {
	if len(data) < 4 {
		return rv, errors.New("Short packet")
	}

	if data[0]>>6 != 1 {
		return rv, errors.New("Invalid version")
	}

	rv.Type = COAPType((data[0] >> 4) & 0x3)
	tokenLen := int(data[0] & 0xf)
	if tokenLen > 8 {
		return rv, ErrInvalidTokenLen
	}

	rv.Code = COAPCode(data[1])
	rv.MessageID = binary.BigEndian.Uint16(data[2:4])

	b := data[4:]
	prev := 0

	for len(b) > 0 {
		if b[0] == 0xff {
			b = b[1:]
			break
		}
		oid := OptionID(prev + int(b[0]>>4))
		l := int(b[0] & 0xf)
		b = b[1:]
		if l == 13 {
			l += int(b[0])
			b = b[1:]
		}
		if l == 14 {
			l = len(b[2:]) & 0xFFFF
			b = b[2:]
		}

		if len(b) < l {
			return rv, errors.New("Truncated")
		}
		var opval interface{} = b[:l]
		switch oid {
		case URIPort, ContentFormat, MaxAge, Accept, Size1:
			opval = decodeInt(b[:l])
		case URIHost, LocationPath, URIPath, URIQuery, LocationQuery,
			ProxyURI, ProxyScheme:
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

	rv.Payload = bytes.Trim(b, " \x00")
	return rv, nil
}
