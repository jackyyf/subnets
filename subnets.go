package subnets

import (
	"net"
)

type bitset []byte
type elem struct {
	val  *matchNode
	next *elem
}

type stack struct {
	*elem
}

func newStack() *stack {
	return &stack{
		elem: new(elem),
	}
}

func (s *stack) Push(val *matchNode) {
	s.elem = &elem{val, s.elem}
}

func (s *stack) Pop() *matchNode {
	var val *matchNode
	val, s.elem = s.elem.val, s.elem.next
	return val
}

func (s bitset) Get(idx int) int {
	bdx := idx >> 3
	bmask := byte(128 >> uint(idx&7))
	if s[bdx]&bmask != 0 {
		return 1
	} else {
		return 0
	}
}

type matchNode struct {
	Full  bool
	Child [2]*matchNode
}

type Matcher struct {
	limit int
	root  *matchNode
}

type IPv4Matcher struct {
	Matcher
}

type IPv6Matcher struct {
	Matcher
}

func Newv4Matcher() *IPv4Matcher {
	return &IPv4Matcher{
		Matcher: Matcher{
			limit: 32,
			root:  new(matchNode),
		},
	}
}

func Newv6Matcher() *IPv6Matcher {
	return &IPv6Matcher{
		Matcher: Matcher{
			limit: 128,
			root:  new(matchNode),
		},
	}
}

func (me *Matcher) Match(ip bitset) bool {
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

func (me *Matcher) Add(ip bitset, plen int) {
	now := me.root
	// Go down to add entry.
	s := newStack()
	for idx := 0; idx < plen; idx++ {
		if now.Full {
			// Do not go further, since the subnet is covered.
			return
		}
		s.Push(now)
		next := ip.Get(idx)
		if now.Child[next] == nil {
			now.Child[next] = new(matchNode)
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
	return me.Matcher.Match(bitset(ipv4))
}

func (me *IPv4Matcher) Add(ipv4 net.IP, plen int) {
	if len(ipv4) != net.IPv4len || plen > net.IPv4len {
		return
	}
	me.Matcher.Add(bitset(ipv4), plen)
}

func (me *IPv6Matcher) Match(ipv6 net.IP) bool {
	if len(ipv6) != net.IPv6len {
		return false
	}
	return me.Matcher.Match(bitset(ipv6))
}

func (me *IPv6Matcher) Add(ipv6 net.IP, plen int) {
	if len(ipv6) != net.IPv6len || plen > net.IPv6len {
		return
	}
	me.Matcher.Add(bitset(ipv6), plen)
}

func (me *IPv4Matcher) AddNet(ipnet *net.IPNet) {
	if ipnet == nil {
		return
	}
	if ones, bits := ipnet.Mask.Size(); len(ipnet.IP) != net.IPv4len || (ones == 0 && bits == 0) || bits != net.IPv4len<<3 {
		return
	} else {
		me.Matcher.Add(bitset(ipnet.IP), ones)
	}
}

func (me *IPv6Matcher) AddNet(ipnet *net.IPNet) {
	if ipnet == nil {
		return
	}
	if ones, bits := ipnet.Mask.Size(); len(ipnet.IP) != net.IPv6len || (ones == 0 && bits == 0) || bits != net.IPv6len<<3 {
		return
	} else {
		me.Matcher.Add(bitset(ipnet.IP), ones)
	}
}
