package coap

import (
	"net"
	"path"
)

type ServeMux struct {
	m map[string]muxEntry
}

type muxEntry struct {
	h       Handler
	pattern string
}

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

// Return the canonical path for p, eliminating . and .. elements.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}
	return np
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

func notFoundHandler(l *net.UDPConn, a *net.UDPAddr, m *Message) *Message {
	return &Message{
		Type: Acknowledgement,
		Code: NotFound,
	}
}

var _ = Handler(&ServeMux{})

func (mux *ServeMux) ServeCOAP(l *net.UDPConn, a *net.UDPAddr, m *Message) *Message {
	h, _ := mux.match(m.PathString())
	if h == nil {
		h, _ = funcHandler(notFoundHandler), ""
	}
	// TODO:  Rewrite path?
	return h.ServeCOAP(l, a, m)
}

func (mux *ServeMux) Handle(pattern string, handler Handler) {
	if pattern == "" {
		panic("http: invalid pattern " + pattern)
	}
	if handler == nil {
		panic("http: nil handler")
	}

	mux.m[pattern] = muxEntry{h: handler, pattern: pattern}
}
