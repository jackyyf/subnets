package subnets

import (
	// "fmt"
	"bufio"
	"bytes"
	//"io"
	"io/ioutil"
	"math/rand"
	"net"
	"strings"
	"testing"
)

const RandSize = 32768

func getChnRoute(b *testing.B) (ret *IPv4Matcher) {
	ret = Newv4Matcher()
	chnroute, err := ioutil.ReadFile("testdata/chnroute.txt")
	if err != nil {
		b.Error("No testdata: chnroute.txt")
	}
	reader := bufio.NewReader(bytes.NewReader(chnroute))
	b.StartTimer()
	for line, err := reader.ReadString('\n'); err == nil; line, err = reader.ReadString('\n') {
		ret.AddNet(parseIPNet(strings.TrimRight(line, "\n")))
	}
	b.StopTimer()
	return
}

func getChnRoute6(b *testing.B) (ret *IPv6Matcher) {
	ret = Newv6Matcher()
	chnroute, err := ioutil.ReadFile("testdata/chnroute-v6.txt")
	if err != nil {
		b.Error("No testdata: chnroute-v6.txt")
	}
	reader := bufio.NewReader(bytes.NewReader(chnroute))
	b.StartTimer()
	for line, err := reader.ReadString('\n'); err == nil; line, err = reader.ReadString('\n') {
		ret.AddNet(parseIPNet(strings.TrimRight(line, "\n")))
	}
	b.StopTimer()
	return
}

func parseIPNet(cidr string) (n *net.IPNet) {
	_, n, _ = net.ParseCIDR(cidr)
	return
}

func TestIPv4Basic(t *testing.T) {
	m := Newv4Matcher()
	subnets := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}
	for _, subnet := range subnets {
		m.AddNet(parseIPNet(subnet))
	}
	t.Log("Try 192.168.0.1 ...")
	if res := m.Match(net.IPv4(192, 168, 0, 1).To4()); !res {
		t.Error("Failed!")
	} else {
		t.Log("OK")
	}
	t.Log("Try 172.23.34.45 ...")
	if res := m.Match(net.IPv4(172, 23, 34, 45).To4()); !res {
		t.Error("Failed!")
	} else {
		t.Log("OK")
	}
	t.Log("Try 10.9.8.7 ...")
	if res := m.Match(net.IPv4(10, 9, 8, 7).To4()); !res {
		t.Error("Failed!")
	} else {
		t.Log("OK")
	}
	t.Log("Try 8.8.8.8 ...")
	if res := m.Match(net.IPv4(8, 8, 8, 8).To4()); res {
		t.Error("Failed!")
	} else {
		t.Log("OK")
	}
	t.Log("Try 1.2.4.8 ...")
	if res := m.Match(net.IPv4(1, 2, 4, 8).To4()); res {
		t.Error("Failed!")
	} else {
		t.Log("OK")
	}
}

func TestIPv4Compress(t *testing.T) {
	m := Newv4Matcher()
	subnets := []string{
		"0.0.0.0/1",
		"128.0.0.0/2",
		"192.0.0.0/2",
	}
	for _, subnet := range subnets {
		m.AddNet(parseIPNet(subnet))
	}
	t.Log("Check compress status...")
	if !m.root.Full {
		t.Error("Error: root is not marked as full!")
	}
	if m.root.Child[0] != nil {
		t.Error("Error: left child is not deleted!")
	}
	if m.root.Child[1] != nil {
		t.Error("Error: right child is not deleted!")
	}
	if !t.Failed() {
		t.Log("OK")
	}
}

func TestIPv6Basic(t *testing.T) {
	m := Newv6Matcher()
	subnets := []string{
		"2001:db8::/64",
		"2001:db8:8000::/48",
		"2001:db8:8001::/48",
	}
	for _, subnet := range subnets {
		m.AddNet(parseIPNet(subnet))
	}
	t.Log("Try 2001:db8::bad:face:f00d:beef ...")
	if res := m.Match(net.ParseIP("2001:db8::bad:face:f00d:beef")); !res {
		t.Error("Failed!")
	} else {
		t.Log("OK")
	}
	t.Log("Try 2001:db8:8000:c0de:bad:face:f00d:beef ...")
	if res := m.Match(net.ParseIP("2001:db8:8000:c0de:bad:face:f00d:beef")); !res {
		t.Error("Failed!")
	} else {
		t.Log("OK")
	}
	t.Log("Try 2001:db8:8001:dead:bad:face:f00d:beef ...")
	if res := m.Match(net.ParseIP("2001:db8:8001:dead:bad:face:f00d:beef")); !res {
		t.Error("Failed!")
	} else {
		t.Log("OK")
	}
	t.Log("Try fe80::1 ...")
	if res := m.Match(net.ParseIP("fe80::1")); res {
		t.Error("Failed!")
	} else {
		t.Log("OK")
	}
	t.Log("Try 2001:db8:: ...")
	if res := m.Match(net.ParseIP("2001:db8:233::")); res {
		t.Error("Failed!")
	} else {
		t.Log("OK")
	}
}

func TestIPv6Compress(t *testing.T) {
	m := Newv6Matcher()
	subnets := []string{
		"::/1",
		"8000::/2",
		"c000::/3",
		"e000::/4",
		"f000::/4",
	}
	for _, subnet := range subnets {
		m.AddNet(parseIPNet(subnet))
	}
	t.Log("Check compress status...")
	if !m.root.Full {
		t.Error("Error: root is not marked as full!")
	}
	if m.root.Child[0] != nil {
		t.Error("Error: left child is not deleted!")
	}
	if m.root.Child[1] != nil {
		t.Error("Error: right child is not deleted!")
	}
	if !t.Failed() {
		t.Log("OK")
	}
}

func randomV4() net.IP {
	data := make([]byte, 4)
	for i := 0; i < 4; i++ {
		data[i] = byte(rand.Int31n(256))
	}
	return net.IP(data)
}

func randomV6() net.IP {
	data := make([]byte, 16)
	for i := 0; i < 16; i++ {
		data[i] = byte(rand.Int31n(256))
	}
	return net.IP(data)
}

func BenchmarkChnRoute(b *testing.B) {
	test_ip4s := make([]net.IP, RandSize)
	for i := 0; i < RandSize; i++ {
		test_ip4s[i] = randomV4()
	}
	b.ResetTimer()
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		r := getChnRoute(b)
		for _, ip4 := range test_ip4s {
			r.Match(ip4)
		}
	}
}

func BenchmarkChnRoute6(b *testing.B) {
	test_ip6s := make([]net.IP, RandSize)
	for i := 0; i < RandSize; i++ {
		test_ip6s[i] = randomV6()
	}
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := getChnRoute6(b)
		for _, ip6 := range test_ip6s {
			r.Match(ip6)
		}
	}
}

func BenchmarkSubnetsGfwIP(b *testing.B) {
	r := getChnRoute(b)
	banfile, err := ioutil.ReadFile("testdata/gfw-fakeip.txt")
	if err != nil {
		b.Error("No testdata: gfw-fakeip.txt")
	}
	reader := bufio.NewReader(bytes.NewReader(banfile))
	banlist := make([]net.IP, 0, 8192)
	for line, err := reader.ReadString('\n'); err == nil; line, err = reader.ReadString('\n') {
		banlist = append(banlist, net.ParseIP(strings.TrimRight(line, "\n")).To4())
	}
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for _, banip := range banlist {
			if r.Match(banip) {
				b.Logf("China IP %s is in gfw fake ip. Correct?", banip.String())
			}
		}
	}
	b.StopTimer()
}

func BenchmarkNaiveGfwIP(b *testing.B) {
	chnroute, err := ioutil.ReadFile("testdata/chnroute.txt")
	if err != nil {
		b.Error("No testdata: chnroute.txt")
	}
	reader := bufio.NewReader(bytes.NewReader(chnroute))
	b.StartTimer()
	nets := make([]*net.IPNet, 0, 8192)
	for line, err := reader.ReadString('\n'); err == nil; line, err = reader.ReadString('\n') {
		net := parseIPNet(strings.TrimRight(line, "\n"))
		if net != nil {
			nets = append(nets, net)
		}
	}
	b.Logf("%d entries in chnroute", len(nets))
	banfile, err := ioutil.ReadFile("testdata/gfw-fakeip.txt")
	if err != nil {
		b.Error("No testdata: gfw-fakeip.txt")
	}
	reader = bufio.NewReader(bytes.NewReader(banfile))
	banlist := make([]net.IP, 0, 8192)
	for line, err := reader.ReadString('\n'); err == nil; line, err = reader.ReadString('\n') {
		banlist = append(banlist, net.ParseIP(strings.TrimRight(line, "\n")).To4())
	}
	b.Logf("%d entries in banlist", len(banlist))
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for _, banip := range banlist {
			for _, net := range nets {
				if net.Contains(banip) {
					b.Logf("China IP %s is in gfw fake ip. Correct?", banip.String())
					break
				}
			}
		}
	}
	b.StopTimer()
}
