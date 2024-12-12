package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProxyTransport(t *testing.T) {
	vpcId, httpProxy, httpsProxy, noProxy := "vpc-xxx", "1.1.1.1", "0.0.0.0", "127.0.0.1"
	//insecure
	transport := NewProxyTransport(vpcId, httpProxy, httpsProxy, noProxy, true)
	proxy, ok := transport.(*proxyTransport)
	assert.True(t, ok)
	assert.Equal(t, proxy.transport, insecureTransport)
	//secure
	transport = NewProxyTransport(vpcId, httpProxy, httpsProxy, noProxy, false)
	proxy, ok = transport.(*proxyTransport)
	assert.True(t, ok)
	assert.Equal(t, proxy.transport, secureTransport)
}
