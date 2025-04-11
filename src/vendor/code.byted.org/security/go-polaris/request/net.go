package request

import (
	"context"
	"errors"
	"net"
	"regexp"
	"strings"

	"code.byted.org/gopkg/env"
	"code.byted.org/security/go-polaris/helpers"
)

// ErrForbiddenAddr 表示请求被封禁时抛出的错误.
//
// ErrForbiddenAddr will be return if the request address is forbidden.
var ErrForbiddenAddr = errors.New("the request address is forbidden")

// ErrSSRFAddr 表示命中了SSRF校验地址
//
// ErrSSRFAddr will be return if the request address is ssrf verify address.
var ErrSSRFAddr = errors.New("the request address is ssrf verify address")

// NetFilter 允许配置网络访问黑白名单规则.
//
// NetFilter allows user to config allowlist/blocklist for network visiting.
type NetFilter interface {
	// Allow 允许访问一个CIDR地址.
	//
	// Allow makes the dialer allow access to a CIDR address.
	Allow(cidr string) error

	// Deny 禁止访问一个CIDR地址. Allow的优先级总是大于Deny.
	//
	// Deny makes the dialer deny access to a CIDR address.
	// The priority of method `Allow` is always greater than `Deny`
	Deny(cidr string) error

	// AllowDomainPort 允许一个带端口的域名，如 "example.com:80"，优先级大于其它黑白名单.
	// AllowDomainPort makes the dialer allow access to a specific domain with
	// port (eg. "example.com:80").
	// The priority of method `AllowDomainPort` is always greater than any others.
	AllowDomainPort(domainWithPort string) error
}

// Dialer 是一个支持黑白名单功能的Dialer.
//
// Dialer is a generic Dialer with allowlist/blocklist support.
type Dialer interface {
	NetFilter

	Dial(network, address string) (net.Conn, error)
	DialContext(ctx context.Context, network, address string) (net.Conn, error)

	// UnderlyingDialer 返回一个底层的`*net.Dialer`，用于调整底层Dialer的配置.
	//
	// UnderlyingDialer returns an underlying `*net.Dialer` for custom
	// configuration.
	UnderlyingDialer() *net.Dialer
}

type dialer struct {
	netDialer           net.Dialer
	ipBlockList         *helpers.CIDRFilter
	ipAllowList         *helpers.CIDRFilter
	domainPortAllowList helpers.PrefixTree
	domainBlockList     helpers.PrefixTree
	subDomainBlockList  []string
}

// Dial implements the Dialer Dial method.
func (d *dialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

// DialContext implements the Dialer DialContext method.
func (d *dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	// check domain port allowlist first
	if d.domainPortAllowList.Find(strings.ToLower(address)) {
		return d.netDialer.DialContext(ctx, network, address)
	}

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	if d.domainBlockList.Find(host) {
		return nil, ErrSSRFAddr
	}

	for _, subdomain := range d.subDomainBlockList {
		if strings.HasSuffix(host, subdomain) {
			return nil, ErrSSRFAddr
		}
	}

	// resolve
	networkType := "ip"
	if env.IsIPV6Only() && !d.isLoopbackIP(host) {
		networkType = "ip6"
	}
	ipAddr, err := net.ResolveIPAddr(networkType, host)
	if err != nil {
		return nil, err
	}

	address = net.JoinHostPort(ipAddr.IP.String(), port)
	if d.ipAllowList.Find(ipAddr.IP) {
		return d.netDialer.DialContext(ctx, network, address)
	}

	if d.ipBlockList.Find(ipAddr.IP) {
		return nil, ErrForbiddenAddr
	}

	return d.netDialer.DialContext(ctx, network, address)
}

func (d *dialer) isLoopbackIP(host string) bool {
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// UnderlyingDialer implements the Dialer UnderlyingDialer method.
func (d *dialer) UnderlyingDialer() *net.Dialer {
	return &d.netDialer
}

// Allow implements the NetFilter Allow method.
func (d *dialer) Allow(cidr string) error {
	return d.ipAllowList.AddCIDR(cidr)
}

// Deny implements the NetFilter Deny method.
func (d *dialer) Deny(cidr string) error {
	return d.ipBlockList.AddCIDR(cidr)
}

const hostPatternStr = `^[a-z0-9\-\.:]+$`
const portPatternStr = `^\d+$`

var hostPattern = regexp.MustCompile(hostPatternStr)
var portPattern = regexp.MustCompile(portPatternStr)

// AllowDomainPort implements the NetFilter AllowDomainPort method.
func (d *dialer) AllowDomainPort(domainWithPort string) error {
	domainWithPort = strings.ToLower(domainWithPort)
	host, port, err := net.SplitHostPort(domainWithPort)
	if err != nil {
		return errors.New("split host port failed: " + err.Error())
	}
	if !hostPattern.MatchString(host) {
		return errors.New("host must match the pattern: " + hostPatternStr)
	}
	if !portPattern.MatchString(port) {
		return errors.New("port must match the pattern: " + portPatternStr)
	}

	d.domainPortAllowList.Add(domainWithPort)
	return nil
}

// NewDialer 创建一个默认禁止访问内部网络的Dialer.
//
// NewDialer create a Dialer and disable local network access by default.
func NewDialer() Dialer {
	return &dialer{
		ipBlockList: helpers.NewLocalNetFilter(),
		ipAllowList: new(helpers.CIDRFilter),
	}
}

// NewBOEDialer 创建一个禁止BOE访问测试域名的Dialer
//
// NewBOEDialer create a Dailer for BOE
func NewBOEDialer() Dialer {
	var domainBlockList helpers.PrefixTree

	// IAST
	domainBlockList.Add("jigsaw-boe.bytedance.net") // boe

	// blackbox-scan
	subDomainBlockList := []string{
		"dast-oast.byte-test.com",
		"oast-cn.byted-dast.com",
		"oast-row.byted-dast.com",
	}

	return &dialer{
		domainBlockList:    domainBlockList,
		ipBlockList:        new(helpers.CIDRFilter),
		ipAllowList:        new(helpers.CIDRFilter),
		subDomainBlockList: subDomainBlockList,
	}
}

// NewInsecureDialer 创建一个允许访问任何网络的Dialer，慎用.
//
// NewInsecureDialer creates a Dialer without any limit, use with caution.
func NewInsecureDialer() Dialer {
	return &dialer{
		ipBlockList: new(helpers.CIDRFilter),
		ipAllowList: new(helpers.CIDRFilter),
	}
}
