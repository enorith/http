package content

import (
	"strings"

	"github.com/enorith/http/contracts"
	"github.com/enorith/supports/str"
)

var (
	trustedProxies         []string
	trustedProxyHeaderSets []string
)

func SetTrustedProxies(proxies ...string) {
	trustedProxies = proxies
}

func SetTrustedProxyHeaderSets(sets ...string) {
	trustedProxyHeaderSets = sets
}

func ExchangeIpFromProxy(addr string, r contracts.RequestContract) string {
	if len(trustedProxies) == 0 || len(trustedProxyHeaderSets) == 0 {
		return addr
	}

	for _, proxy := range trustedProxies {
		if addr == proxy {
			for _, set := range trustedProxyHeaderSets {
				if header := r.HeaderString(set); header != "" {
					return resolveForwardHeader(header)
				}
			}
		}
	}

	return addr
}

func resolveForwardHeader(forward string) string {
	if str.Contains(forward, ",") {
		return strings.Split(forward, ",")[0]
	}

	return forward
}
