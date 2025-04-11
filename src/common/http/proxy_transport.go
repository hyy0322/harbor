package http

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"code.byted.org/security/go-polaris/request"
	"golang.org/x/net/http/httpproxy"
)

const (
	httpProxy  requestKey = "http.proxy"
	httpsProxy requestKey = "https.proxy"
	noProxy    requestKey = "http.noProxy"
	vpcId      requestKey = "http.vpcId"

	vpcIdHeader string = "X-Target-VPC"
)

type requestKey string

var (
	insecureTransport = withInsecure(newTransportForProxy())
	secureTransport   = withSecure(newTransportForProxy())

	insecureTransportNoProxy = withInsecure(newTransportNoProxy())
	secureTransportNoProxy   = withSecure(newTransportNoProxy())
)

type proxyTransport struct {
	transport  http.RoundTripper
	httpProxy  string
	httpsProxy string
	noProxy    string
	vpcId      string
}

func (p *proxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := context.WithValue(req.Context(), httpProxy, p.httpProxy)
	ctx = context.WithValue(ctx, httpsProxy, p.httpsProxy)
	ctx = context.WithValue(ctx, noProxy, p.noProxy)
	ctx = context.WithValue(ctx, vpcId, p.vpcId)
	req = req.WithContext(ctx)
	req.Header.Set(vpcIdHeader, p.vpcId)
	return p.transport.RoundTrip(req)
}

// NewProxyTransport ...
func NewProxyTransport(vpcId, httpProxy, httpsProxy, noProxy string, insecure bool) http.RoundTripper {
	var underlayTransport http.RoundTripper
	if httpProxy == "" && httpsProxy == "" {
		underlayTransport = getTransportNoProxy(insecure)
	} else {
		underlayTransport = getTransportForProxy(insecure)
	}
	return &proxyTransport{
		transport:  underlayTransport,
		httpProxy:  httpProxy,
		httpsProxy: httpsProxy,
		noProxy:    noProxy,
		vpcId:      vpcId,
	}
}

func getTransportForProxy(insecure bool) *http.Transport {
	if insecure {
		return insecureTransport
	}
	return secureTransport
}

func getTransportNoProxy(insecure bool) *http.Transport {
	if insecure {
		return insecureTransportNoProxy
	}
	return secureTransportNoProxy
}

func newTransportForProxy() *http.Transport {
	return &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			httpProxy, ok := req.Context().Value(httpProxy).(string)
			if !ok {
				return nil, nil
			}
			httpsProxy, ok := req.Context().Value(httpsProxy).(string)
			if !ok {
				return nil, nil
			}
			noProxy, ok := req.Context().Value(vpcId).(string)
			if !ok {
				return nil, nil
			}
			return (&httpproxy.Config{
				HTTPProxy:  httpProxy,
				HTTPSProxy: httpsProxy,
				NoProxy:    noProxy,
			}).ProxyFunc()(req.URL)
		},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		// https场景，需要在Connect接口中先携带targetvpc信息，因此需要在Connect接口中增加Header
		GetProxyConnectHeader: func(ctx context.Context, proxyUrl *url.URL, target string) (header http.Header, err error) {
			// ctx 来自于req，所以可以直接获取
			vpcID, ok := ctx.Value(vpcId).(string)
			if !ok {
				return nil, nil
			}
			return http.Header{
				vpcIdHeader: []string{vpcID},
			}, nil
		},
		TLSClientConfig: &tls.Config{},
		// 在最终模式下，会出现走相同的代理、相同目的IP，但是vpc不同的场景。keepalive模式下会复用相同proxy、目的地址的连接，导致转发到错误的vpc，所以需要关闭keepalive
		DisableKeepAlives:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

type Dialer struct {
	request.Dialer
	subDomainAllowList []string
}

func (d *Dialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	for _, subdomain := range d.subDomainAllowList {
		if strings.HasSuffix(host, subdomain) {
			return d.Dialer.UnderlyingDialer().DialContext(ctx, network, addr)
		}
	}

	return d.Dialer.DialContext(ctx, network, addr)
}

func newTransportNoProxy() *http.Transport {
	dailer := request.NewInsecureDialer()
	dailer.Deny("100.96.0.96/32")
	dailer.UnderlyingDialer().Timeout = 30 * time.Second
	dailer.UnderlyingDialer().KeepAlive = 30 * time.Second
	dailer.UnderlyingDialer().DualStack = true

	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dailer.DialContext,
		TLSClientConfig:       &tls.Config{},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func withInsecure(t *http.Transport) *http.Transport {
	t.TLSClientConfig.InsecureSkipVerify = true
	return t
}

func withSecure(t *http.Transport) *http.Transport {
	if InternalTLSEnabled() {
		tlsConfig, err := GetInternalTLSConfig()
		if err != nil {
			panic(err)
		}
		t.TLSClientConfig = tlsConfig
	}
	return t
}
