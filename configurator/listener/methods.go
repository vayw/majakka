package main

import (
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
)

func (l *Listener) SetListenerFilters() {
	// SetListenerFilters adds listener filters according to listener port
	// tls_inspector for port 443, for example
	switch l.Port {
	case 443:
		l.ListenerFilters = []string{"envoy.filters.listener.tls_inspector", "envoy.filters.listener.http_inspector"}
	case 80:
		l.ListenerFilters = []string{"envoy.filters.listener.http_inspector"}
	default:
		l.ListenerFilters = []string{}
	}
}

func (l Listener) Generate() *listener.Listener {
	nlistener := &listener.Listener{
		Name: l.Name,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  l.Address,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: l.Port,
					},
				},
			},
		},
	}
	for _, fltr := range l.ListenerFilters {
		nlistener.ListenerFilters = append(nlistener.ListenerFilters, &listener.ListenerFilter{Name: fltr})
	}
	//		FilterChains: []*listener.FilterChain{{
	//			Filters: []*listener.Filter{{
	//				Name: wellknown.HTTPConnectionManager,
	//				ConfigType: &listener.Filter_TypedConfig{
	//					TypedConfig: pbst,
	//				},
	//			}},
	//		}},
	//	}
}
