package helpers

import (
	"bytes"
	"fmt"
	"net"

	"code.byted.org/security/go-polaris/report"
)

var reporter = report.NewReporter("helpers")

type cidr struct {
	min []byte
	max []byte
}

func (x *cidr) String() string {
	return fmt.Sprintf("{%v, %v}", x.min, x.max)
}

func (x *cidr) compare(y *cidr) int {
	if bytes.Compare(x.max, y.min) < 0 {
		return -1
	}
	if bytes.Compare(x.min, y.max) > 0 {
		return 1
	}
	return 0
}

func (x *cidr) contains(y *cidr) bool {
	if bytes.Compare(x.min, y.min) > 0 {
		return false
	}
	if bytes.Compare(x.max, y.max) < 0 {
		return false
	}

	return true
}

type avlNode struct {
	value  *cidr
	parent *avlNode
	left   *avlNode
	right  *avlNode
	h      int
}

func (x *avlNode) height() int {
	if x == nil {
		return -1
	}

	return x.h
}

func (x *avlNode) updateHeight() {
	x.h = x.left.height() + 1
	rh := x.right.height() + 1
	if x.h < rh {
		x.h = rh
	}
}

func (x *avlNode) rotateLeft() *avlNode {
	z := x.right
	t23 := z.left
	x.right = t23
	if t23 != nil {
		t23.parent = x
	}
	z.left = x
	x.parent = z

	x.updateHeight()
	z.updateHeight()
	return z
}

func (x *avlNode) rotateRight() *avlNode {
	z := x.left
	t23 := z.right
	x.left = t23
	if t23 != nil {
		t23.parent = x
	}
	z.right = x
	x.parent = z

	x.updateHeight()
	z.updateHeight()
	return z
}

func (x *avlNode) rotateRightLeft() *avlNode {
	x.right = x.right.rotateRight()
	return x.rotateLeft()
}

func (x *avlNode) rotateLeftRight() *avlNode {
	x.left = x.left.rotateLeft()
	return x.rotateRight()
}

// rebalance from current node and return new root.
func (x *avlNode) rebalance() *avlNode {
	var p *avlNode
	for p = x.parent; p != nil; x, p = p, p.parent {
		grandParent := p.parent
		isLeft := grandParent != nil && grandParent.left == p
		oldHeight := p.height()

		balance := p.right.height() - p.left.height()
		if balance > 1 { // right-heavy
			if x.left.height() > x.right.height() {
				p = p.rotateRightLeft()
			} else {
				p = p.rotateLeft()
			}
			p.parent = grandParent
		} else if balance < -1 { // left-heavy
			if x.right.height() > x.left.height() {
				p = p.rotateLeftRight()
			} else {
				p = p.rotateRight()
			}
			p.parent = grandParent
		} else { // still balanced
			p.updateHeight()
		}

		if grandParent == nil {
			return p
		}

		if isLeft {
			grandParent.left = p
		} else {
			grandParent.right = p
		}

		if oldHeight == p.height() {
			break
		}
	}

	for p.parent != nil {
		p = p.parent
	}

	return p
}

// insert and return new root.
func (x *avlNode) insert(z *avlNode) *avlNode {
	if x == nil {
		return z
	}
	root := x

	cnt := true // continue
	for cnt {
		switch x.value.compare(z.value) {
		case 0: // equal
			if !x.value.contains(z.value) {
				x.value = z.value
			}
			return root
		case -1: // x < z
			if x.right == nil {
				x.right = z
				z.parent = x
				cnt = false
			}
			x = x.right
		case 1: // x > z
			if x.left == nil {
				x.left = z
				z.parent = x
				cnt = false
			}
			x = x.left
		}
	}

	return z.rebalance()
}

func (x *avlNode) find(ip net.IP) bool {
	ipBytes := []byte(ip)
	v := &cidr{ipBytes, ipBytes}

	for x != nil {
		cmp := x.value.compare(v)
		if cmp < 0 {
			x = x.right
		} else if cmp > 0 {
			x = x.left
		} else if x.value.contains(v) {
			return true
		}
	}

	return false
}

// CIDRFilter 是一个用AVL实现的高性能IP地址过滤器.
//
// CIDRFilter is an AVL-based, high performance IP filter.
type CIDRFilter struct {
	root *avlNode
}

func genMaxIP(ipnet *net.IPNet) []byte {
	ip := []byte(ipnet.IP)
	ret := make([]byte, len(ip))

	mask := []byte(ipnet.Mask)

	for i := 0; i < len(ip); i++ {
		ret[i] = ip[i] | (^mask[i])
	}

	return ret
}

// ipNetToIPv6Len ipnet with length of IPv6len.
func ipNetToIPv6Len(ipnet *net.IPNet) {
	if len(ipnet.Mask) == net.IPv6len {
		return
	}

	newMask := make([]byte, net.IPv6len)
	for i := 0; i < 12; i++ {
		newMask[i] = 0xff
	}
	copy(newMask[12:], ipnet.Mask)
	ipnet.Mask = newMask
	ipnet.IP = ipnet.IP.To16()
}

// AddCIDR 用于给过滤器添加一个CIDR地址，如"192.168.0.0/16" / "fe80::/10" .
//
// AddCIDR adds a CIDR address (eg. "192.168.0.0/16" or "fe80::/10") into the filter,
func (c *CIDRFilter) AddCIDR(s string) error {
	_, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		return err
	}

	// ensure len(IP) == IPv6len
	ipNetToIPv6Len(ipnet)

	min := []byte(ipnet.IP)
	max := genMaxIP(ipnet)

	v := &cidr{min, max}

	c.root = c.root.insert(&avlNode{value: v})
	return nil
}

// Find 用于检查一个IP地址是否在当前过滤器中.
//
// Find checks if the `ip` satisfies the filter.
func (c *CIDRFilter) Find(ip net.IP) bool {
	return c.root.find(ip.To16())
}

// NewLocalNetFilter 创建一个用于过滤本地网络地址的Filter.
// 由于所有IPv4地址都会被转成IPv6地址，因此，不需要单独考虑 IPv4-mapped IPv6 address 的情况.
//
// NewLocalNetFilter creates a filter and add local network into the filter.
// No need to care about IPv4-mapped IPv6 address.
func NewLocalNetFilter() *CIDRFilter {
	filter := new(CIDRFilter)
	for _, cidr := range []string{
		"0.0.0.0/8",     // self
		"10.0.0.0/8",    // RFC1918
		"100.64.0.0/10", // RFC6598 https://datatracker.ietf.org/doc/html/rfc6598
		"127.0.0.0/8",   // IPv4 loopback
		"169.254.0.0/16",
		"172.16.0.0/12", // RFC1918
		"192.0.0.0/29",
		"192.0.0.170/31",
		"192.0.2.0/24",
		"192.18.0.0/15",
		"192.168.0.0/16", // RFC1918
		"192.51.100.0/24",
		"203.0.113.0/24",
		"240.0.0.0/4",
		"255.255.255.255/32",
		"33.0.0.0/8", // ToB机房启用“公网保留”ipv4地址33.0.0.0/8; ToB IDC reserved ipv4 address 33.0.0.0/8
		"::/128",     // self
		"::1/128",    // IPv6 loopback
		// "::ffff:0:0/96", // IPv4-mapped addresses [RFC4291]
		"100::/64",
		"2001::/23",
		"2001:2::/48",
		"2001:db8::/32",
		"2001:10::/28",
		"fe80::/10", // IPv6 link-local
		"fc00::/7",  // IPv6 unique local addr

		// internal doc: https://tech.bytedance.net/articles/6934480248100618247?from=net_app_search#heading31
		"fdbd::/16",
		"2605:340:cd00::/40",
		"2600:2d00::/28",
	} {
		_ = filter.AddCIDR(cidr)
	}

	return filter
}
