package coap

// ServeMux provides mappings from a common endpoint to handlers by
// request path.
type ServeMux struct {
	m map[string]muxEntry
}

type muxEntry struct {
	h       Handler
	pattern string
}

var DefaultServeMux = NewServeMux()

// NewServeMux creates a new ServeMux.
func NewServeMux() *ServeMux { return &ServeMux{m: make(map[string]muxEntry)} }

// Does path match pattern?
func pathMatch(pattern, path string) bool {
	if len(pattern) == 0 {
		// should not happen
		return false
	}
	n := len(pattern)
	if pattern[n-1] != '/' {
		return pattern == path
	}
	return len(path) >= n && path[0:n] == pattern
}

// Find a handler on a handler map given a path string
// Most-specific (longest) pattern wins
func (mux *ServeMux) match(path string) (h Handler, pattern string) {
	var n = 0
	for k, v := range mux.m {
		if !pathMatch(k, path) {
			continue
		}
		if h == nil || len(k) > n {
			n = len(k)
			h = v.h
			pattern = v.pattern
		}
	}
	return
}

// ServeCOAP handles a single COAP message.  The message arrives from
// the given listener having originated from the given UDPAddr.
func (mux *ServeMux) ServeCOAP(r *RemoteAddr, m *Message) *Message {
	h, _ := mux.match(m.PathString())
	return h.ServeCOAP(r, m)
}

// Handle configures a handler for the given path.
func (mux *ServeMux) Handle(pattern string, handler Handler) {
	for pattern != "" && pattern[0] == '/' {
		pattern = pattern[1:]
	}

	if pattern == "" {
		panic("coap: invalid pattern " + pattern)
	}

	if handler == nil {
		panic("coap: nil handler")
	}

	if handler == nil {
		debugMsg("coap: nil handler")
		return
	}

	mux.m[pattern] = muxEntry{h: handler, pattern: pattern}
}

// HandleFunc configures a handler for the given path.
func (mux *ServeMux) HandleFunc(pattern string, handler func(r *RemoteAddr, m *Message) *Message) {
	mux.Handle(pattern, HandlerFunc(handler))
}
