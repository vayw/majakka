package main

import (
	"errors"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
)

const StateEnabled = "enabled"
const StateDisabled = "disabled"

type ClustersMap map[string]*Cluster
type RouteConfMap map[string]*RouteConf
type ListenersMap map[string]*Listener
type EndpointsMap map[string]*Endpoint
type RouteAssigments map[string]bool
type VHostRoutes map[string]Route
type VHosts map[string]VHost

type VHost struct {
	Name        string
	Type        string
	Domains     []string
	TLSOnly     bool
	Routes      VHostRoutes
	TLSCertPath string
	TLSKeyPath  string
}

type Route struct {
	Prefix  string
	Headers map[string]string
	Cluster string
	Mirrors map[string]uint32
}

type Endpoint struct {
	UpstreamHost string
	UpstreamPort uint32
	State        string
}

type Cluster struct {
	Name      string
	Endpoints EndpointsMap
}

type Configuration struct {
	Clusters      ClustersMap
	RouteConf     RouteConfMap
	Listeners     ListenersMap
	VirtualHosts  VHosts
	SnapshotCache *cache.SnapshotCache
}

func (cf *Configuration) AddCluster(name string) error {
	if _, ok := cf.Clusters[name]; ok {
		return errors.New("Cluster already exists")
	} else {
		cf.Clusters[name] = &Cluster{name, make(EndpointsMap)}
		return nil
	}
}

func (cf *Configuration) AddEndpoint(name, cluster, address string, port uint32) error {
	if _, ok := cf.Clusters[cluster]; ok {
		cf.Clusters[cluster].Endpoints[name] = &Endpoint{address, port, StateEnabled}
	} else {
		_ = cf.AddCluster(cluster)
		cf.Clusters[cluster].Endpoints[name] = &Endpoint{address, port, StateEnabled}
	}
	err := cf.GenerateSnapshot()
	return err
}

func (cf *Configuration) CheckEndpoint(name, cluster string) error {
	if _, ok := cf.Clusters[cluster]; ok {
		if _, ok = cf.Clusters[cluster].Endpoints[name]; ok {
			return nil
		} else {
			return errors.New("Endpoint not found")
		}
	} else {
		return errors.New("Cluster not found")
	}
}

func (cf *Configuration) DeleteEndpoint(name, cluster string) error {
	err := cf.CheckEndpoint(name, cluster)
	if err != nil {
		return err
	} else {
		delete(cf.Clusters[cluster].Endpoints, name)
		err := cf.GenerateSnapshot()
		return err
	}
}

func (cf *Configuration) DisableEndpoint(name, cluster string) error {
	err := cf.CheckEndpoint(name, cluster)
	if err != nil {
		return err
	} else {
		cf.Clusters[cluster].Endpoints[name].State = StateDisabled
		err := cf.GenerateSnapshot()
		return err
	}
}

func (cf *Configuration) EnableEndpoint(name, cluster string) error {
	err := cf.CheckEndpoint(name, cluster)
	if err != nil {
		return err
	} else {
		cf.Clusters[cluster].Endpoints[name].State = StateEnabled
		err := cf.GenerateSnapshot()
		return err
	}
}

func (cf *Configuration) AddRouteConf(name string) error {
	if _, ok := cf.RouteConf[name]; ok {
		return errors.New("Route already exists")
	} else {
		cf.RouteConf[name] = &RouteConf{
			Name:       name,
			Assigments: make(RouteAssigments),
		}
		cf.ProcessListenerStates(name)
		err := cf.GenerateSnapshot()
		return err
	}
}

func (cf *Configuration) RouteConfUsable(name string) bool {
	if _, ok := cf.RouteConf[name]; ok {
		if len(cf.RouteConf[name].VHosts) > 0 {
			return true
		}
	}
	return false
}

func (cf *Configuration) RouteAssign(route, listener string) error {
	cf.RouteConf[route].Assigments[listener] = true
	return nil
}

func (cf *Configuration) AddListener(name, address string, port uint32, route string) error {
	if _, ok := cf.Listeners[name]; ok {
		return errors.New("Listener already exists")
	} else {
		if cf.RouteConfUsable(route) {
			cf.Listeners[name] = &Listener{name, address, port, route, StateEnabled}
			_ = cf.RouteAssign(route, name)
		} else {
			cf.Listeners[name] = &Listener{name, address, port, route, StateDisabled}
		}
		err := cf.GenerateSnapshot()
		return err
	}
}

func (cf *Configuration) ProcessListenerStates(route string) {
	for _, l := range cf.Listeners {
		if l.Route == route && l.State == StateDisabled && cf.RouteConfUsable(l.Route) {
			l.State = StateEnabled
			cf.RouteConf[route].Assigments[l.Name] = true
		}
	}
}

func (cf *Configuration) AddMirroring(routeconf, vhost, route, cluster string, fraction uint32) error {
	cf.RouteConf[routeconf].VHosts[vhost].Routes[route].Mirrors[cluster] = fraction
	err := cf.GenerateSnapshot()
	return err
}

func (cf *Configuration) DeleteMirroring(routeconf, vhost, route, cluster string) error {
	delete(cf.RouteConf[routeconf].VHosts[vhost].Routes[route].Mirrors, cluster)
	err := cf.GenerateSnapshot()
	return err
}

func (cf *Configuration) AddVHost(name, vhosttype string) error {
	if err, ok := cf.VirtualHosts[name]; !ok {
		cf.VirtualHosts[name] = VHost{
			Name: name,
			Type: vhosttype,
		}
		err := cf.GenerateSnapshot()
	}
	return err
}
