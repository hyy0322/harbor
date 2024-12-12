package http

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

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
	return &proxyTransport{
		transport:  getTransportForProxy(insecure),
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
