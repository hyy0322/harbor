// Package request 用于提供一种可控制访问范围的请求方式.
//
// 该包的初衷是提供一种非侵入式的实现，解决公司内部频发的SSRF问题。
// 由于SSRF的防御是在传输层完成的，开发者可无需考虑 301/302/307 等重定向，也无需考虑
// DNS rebinding攻击。
//
// Eg 1. 禁止访问内网。
// Prevent Go from accessing the intranet
//
//	import (
//		"code.byted.org/security/go-polaris/request"
//		"fmt"
//	)
//
//	func main() {
//		client := request.NewClient()
//		_, err := client.Get("https://example.com/")
//		fmt.Println(err) // <nil>
//
//		_, err = client.Get("https://wiki.bytedance.net/")
//		fmt.Println(err) // Get https://wiki.bytedance.net/: the request address is forbidden
//	}
//
// Eg 2. 禁止访问内网，但放行特定域名:端口。
// Prevent Go from accessing the intranet, but allow access to specific intranet services.
//
//	import (
//		"code.byted.org/security/go-polaris/request"
//		"fmt"
//	)
//
//	func main() {
//		transport := request.NewTransport()
//		transport.AllowDomainPort("wiki.bytedance.net:443")
//		client := request.NewClientWithTransport(transport)
//
//		_, err := client.Get("https://wiki.bytedance.net/")
//		fmt.Println(err) // <nil>
//
//		_, err = client.Get("https://code.byted.org/")
//		fmt.Println(err) // Get https://code.byted.org/: the request address is forbidden
//	}
//
// 常见问题1: 为何访问的并非是内网地址，还是返回`the request address is forbidden`?
//
// 最常见的原因是运行环境中配置了代理，http.Client建立连接前首先会和代理创建TCP连接。
// 如果此时，代理的IP为内网IP，那么请求将无法建立。请检查运行环境的 `https_proxy` 或
// `http_proxy` 环境变量，确保未配置任何代理。
//
// 常见问题2: 如果必须使用代理怎么办？
//
// 很遗憾，由于抵御SSRF是在传输层层面做的，使用代理相当于绕过了这种抵御方式，导致 request 包
// 提供的黑白名单策略完全失效。理论上，request包和代理机制不兼容。
package request

import (
	"net"
	"net/http"
	"time"

	"code.byted.org/gopkg/env"
)

// Option 操作选项
type Option func(*http.Client) *http.Client

// 相关Client选项
var (
	// OmitBOE BOE环境下豁免内网访问限制
	// if detect boe, exempt from intranet access limiation
	OmitBOE Option = func(c *http.Client) *http.Client {
		if env.IsBoe() {
			c.Transport = NewBOETransport().RoundTripper()
		}
		return c
	}
)

// Transport 是一个自带访问控制的 http.Transport.
//
// Transport is a secured version of http.Transport.
type Transport interface {
	NetFilter

	// Underlyings 返回底层的`*http.Transport`与`*net.Dialer`，只用于调整配置，
	// 切勿直接使用。
	//
	// Underlyings returns an underlying `*http.Transport` and `*net.Dialer`
	// for custom configuration ONLY. Use with caution.
	Underlyings() (*http.Transport, *net.Dialer)

	// RoundTripper returns a `http.RoundTripper`.
	RoundTripper() http.RoundTripper
}

type transport struct {
	Dialer
	httpTransport *http.Transport
}

// Underlyings implements the Transport Underlyings method.
func (t *transport) Underlyings() (*http.Transport, *net.Dialer) {
	return t.httpTransport, t.Dialer.UnderlyingDialer()
}

// RoundTripper implements the Transport RoundTripper method.
func (t *transport) RoundTripper() http.RoundTripper {
	// make sure the dialContext has been replaced
	t.httpTransport.DialContext = t.Dialer.DialContext
	t.httpTransport.Dial = t.Dialer.Dial // nolint

	return t.httpTransport
}

// NewTransport 创建一个新的Transport，禁止本地网络访问.
//
// NewTransport create a new Transport and disable local network access by default.
func NewTransport() Transport {
	dialer := NewDialer()
	netDialer := dialer.UnderlyingDialer()
	netDialer.Timeout = 30 * time.Second
	netDialer.KeepAlive = 30 * time.Second

	return NewTransportWithDialer(dialer)
}

// NewBOETransport 创建一个供BOE环境使用的Transport
//
// NewBOETransport create a new Transport for BOE
func NewBOETransport() Transport {
	dialer := NewBOEDialer()
	netDialer := dialer.UnderlyingDialer()
	netDialer.Timeout = 30 * time.Second
	netDialer.KeepAlive = 30 * time.Second

	return NewTransportWithDialer(dialer)
}

// NewTransportWithDialer 根据 `dialer` 新建一个Transport.
//
// NewTransportWithDialer creates a new Transport with `dialer`.
func NewTransportWithDialer(dialer Dialer) Transport {
	return &transport{
		Dialer:        dialer,
		httpTransport: makeDefaultTransport(),
	}
}

// NewTransportWithHTTPTransport 根据 `dialer` 和原生 `httpTransport` 新建一个Transport
// 传入的 http.Transport 会被修改，不应该继续使用
//
// NewTransportWithHTTPTransport creates a new Transport with `dialer` and raw `httpTransport`.
// the given http.Transport WILL be modified and should not be reused.
func NewTransportWithHTTPTransport(dialer Dialer, httpTransport *http.Transport) Transport {
	httpTransport.ForceAttemptHTTP2 = true
	return &transport{
		Dialer:        dialer,
		httpTransport: httpTransport,
	}
}

// NewClient 创建一个禁止访问内网的`*http.Client`.
// 创建后切勿修改 Client 的 Transport 属性，否则会导致内网访问控制失效.
// 如果需要自定义 Transport，请改用 NewTransport + NewClientWithTransport 方法.
//
// NewClient creates a standard `*http.Client` and disable local network access by default.
// Please do NOT change the value of Client.Transport, or the ACL will NOT work properly.
// If you need to custom the Transport settings, plz use NewTransport with NewClientWithTransport.
func NewClient(opts ...Option) *http.Client {
	client := NewClientWithTransport(NewTransport())
	for _, opt := range opts {
		client = opt(client)
	}
	return client
}

// NewClientWithTransport 根据提供的`tranport`信息创建一个`*http.Client`.
// 用于满足自定义白名单的功能.
//
// NewClientWithTransport creates a standard `*http.Client` with provided `transport`.
// `transport` allows user to config allowlist/blocklist.
func NewClientWithTransport(transport Transport) *http.Client {
	return &http.Client{
		Transport: transport.RoundTripper(),
	}
}
