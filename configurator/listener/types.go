package main

type Listener struct {
	Name            string
	Address         string
	Port            uint32
	ListenerFilters []string
	FilterChains    []string
	State           string
}
