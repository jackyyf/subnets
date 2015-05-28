package subnets

import (
	"net"
)

type bitset []byte

type stack struct {
	data []*matchNode
	idx  int
}

func newStack(size int) *stack {
	return &stack{
		data: make([]*matchNode, size),
		idx:  -1,
	}
}

func (s *stack) Push(val *matchNode) {
	s.idx++
	s.data[s.idx] = val
}

func (s *stack) Pop() (ret *matchNode) {
	ret = s.data[s.idx]
	s.idx--
	return
}

func (s *stack) Clear() {
	s.idx = -1
}

func (s bitset) Get(idx int) int {
	bdx := idx >> 3
	bmask := byte(128 >> uint(idx&7))
	if s[bdx]&bmask != 0 {
		return 1
	}
	return 0
}

type matchNode struct {
	Full  bool
	Child [2]*matchNode
}

type matcher struct {
	limit int
	root  *matchNode
	stack *stack
}

type IPv4Matcher struct {
	matcher
}

type IPv6Matcher struct {
	matcher
}

var pool []matchNode

func newNode() (ret *matchNode) {
	if len(pool) == 0 {
		pool = make([]matchNode, 1024)
	}
	ret = &pool[0]
	pool = pool[1:]
	return
}

func Newv4Matcher() *IPv4Matcher {
	return &IPv4Matcher{
		matcher: matcher{
			limit: 32,
			root:  newNode(),
		},
	}
}

func Newv6Matcher() *IPv6Matcher {
	return &IPv6Matcher{
		matcher: matcher{
			limit: 128,
			root:  newNode(),
		},
	}
}

func (me *matcher) Match(ip bitset) bool {
	now := me.root
	for idx := 0; idx < me.limit; idx++ {
		if now.Full {
			return true
		}
		next := ip.Get(idx)
		if now.Child[next] == nil {
			return false
		}
		now = now.Child[next]
	}
	return false
}

func (me *matcher) Add(ip bitset, plen int) {
	now := me.root
	// Go down to add entry.
	if me.stack == nil {
		me.stack = newStack(me.limit)
	}
	s := me.stack
	s.Clear()
	for idx := 0; idx < plen; idx++ {
		if now.Full {
			// Do not go further, since the subnet is covered.
			return
		}
		s.Push(now)
		next := ip.Get(idx)
		if now.Child[next] == nil {
			now.Child[next] = newNode()
		}
		now = now.Child[next]
	}
	now.Full = true
	for idx := 0; idx < plen; idx++ {
		now = s.Pop()
		if (now.Child[0] != nil && now.Child[0].Full) &&
			(now.Child[1] != nil && now.Child[1].Full) {
			now.Full = true
			now.Child[0], now.Child[1] = nil, nil
		} else {
			return
		}
	}
}

func (me *IPv4Matcher) Match(ipv4 net.IP) bool {
	if len(ipv4) != net.IPv4len {
		return false
	}
	return me.matcher.Match(bitset(ipv4))
}

func (me *IPv4Matcher) Add(ipv4 net.IP, plen int) {
	if len(ipv4) != net.IPv4len || plen > net.IPv4len<<3 {
		return
	}
	me.matcher.Add(bitset(ipv4), plen)
}

func (me *IPv6Matcher) Match(ipv6 net.IP) bool {
	if len(ipv6) != net.IPv6len {
		return false
	}
	return me.matcher.Match(bitset(ipv6))
}

func (me *IPv6Matcher) Add(ipv6 net.IP, plen int) {
	if len(ipv6) != net.IPv6len || plen > net.IPv6len<<3 {
		return
	}
	me.matcher.Add(bitset(ipv6), plen)
}

func (me *IPv4Matcher) AddNet(ipnet *net.IPNet) {
	if ipnet == nil {
		return
	}
	if ones, bits := ipnet.Mask.Size(); len(ipnet.IP) != net.IPv4len || (ones == 0 && bits == 0) || bits != net.IPv4len<<3 {
		return
	} else {
		me.matcher.Add(bitset(ipnet.IP), ones)
	}
}

func (me *IPv6Matcher) AddNet(ipnet *net.IPNet) {
	if ipnet == nil {
		return
	}
	if ones, bits := ipnet.Mask.Size(); len(ipnet.IP) != net.IPv6len || (ones == 0 && bits == 0) || bits != net.IPv6len<<3 {
		return
	} else {
		me.matcher.Add(bitset(ipnet.IP), ones)
	}
}
