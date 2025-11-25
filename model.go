package regexp

import (
	"regexp"
	"regexp/syntax"
	"sync"
)

type Regexp = regexp.Regexp

var l *sync.RWMutex

type record struct {
	Re   string
	Mode syntax.Flags
}

// cached saves all compiled regexp here
var cached = make(map[record]*Regexp)
