package main

import (
	"errors"
	"strconv"
	"time"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
)

const StateEnabled = "enabled"
const StateDisabled = "disabled"

type ClustersMap map[string]*Cluster
type RouteConfMap map[string]*RouteConf
type ListenersMap map[string]*Listener
type EndpointsMap map[string]*Endpoint
type RouteAssigments map[string]bool
type Mirrors map[string]uint32
type VirtualHosts map[string]*VHost

type VHost struct {
	Name    string
	Domains []string
	Routes  []route.Route
}

type Listener struct {
	Name    string
	Address string
	Port    uint32
	Route   string
	State   string
}

type RouteConf struct {
	Name       string
	Assigments RouteAssigments
	Mirroring  Mirrors
	Cluster    string
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
	SnapshotCache *cache.SnapshotCache
}

func (cf Configuration) AddCluster(name string) error {
	if _, ok := cf.Clusters[name]; ok {
		return errors.New("Cluster already exists")
	} else {
		cf.Clusters[name] = &Cluster{name, make(EndpointsMap)}
		return nil
	}
}

func (cf Configuration) AddEndpoint(name, cluster, address string, port uint32) error {
	if _, ok := cf.Clusters[cluster]; ok {
		cf.Clusters[cluster].Endpoints[name] = &Endpoint{address, port, StateEnabled}
	} else {
		_ = cf.AddCluster(cluster)
		cf.Clusters[cluster].Endpoints[name] = &Endpoint{address, port, StateEnabled}
	}
	err := cf.GenerateSnapshot()
	return err
}

func (cf Configuration) CheckEndpoint(name, cluster string) error {
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

func (cf Configuration) DeleteEndpoint(name, cluster string) error {
	err := cf.CheckEndpoint(name, cluster)
	if err != nil {
		return err
	} else {
		delete(cf.Clusters[cluster].Endpoints, name)
		err := cf.GenerateSnapshot()
		return err
	}
}

func (cf Configuration) DisableEndpoint(name, cluster string) error {
	err := cf.CheckEndpoint(name, cluster)
	if err != nil {
		return err
	} else {
		cf.Clusters[cluster].Endpoints[name].State = StateDisabled
		err := cf.GenerateSnapshot()
		return err
	}
}

func (cf Configuration) EnableEndpoint(name, cluster string) error {
	err := cf.CheckEndpoint(name, cluster)
	if err != nil {
		return err
	} else {
		cf.Clusters[cluster].Endpoints[name].State = StateEnabled
		err := cf.GenerateSnapshot()
		return err
	}
}

func (cf Configuration) AddRoute(name, cluster string) error {
	if _, ok := cf.RouteConf[name]; ok {
		return errors.New("Route already exists")
	} else {
		cf.RouteConf[name] = &RouteConf{
			Name:       name,
			Assigments: make(RouteAssigments),
			Cluster:    cluster,
			Mirroring:  make(Mirrors),
		}
		cf.ListenerCheck(name)
		err := cf.GenerateSnapshot()
		return err
	}
}

func (cf Configuration) RouteOk(name string) bool {
	if _, ok := cf.RouteConf[name]; ok {
		return true
	} else {
		return false
	}
}

func (cf Configuration) RouteAssign(route, listener string) error {
	cf.RouteConf[route].Assigments[listener] = true
	return nil
}

func (cf Configuration) AddListener(name, address string, port uint32, route string) error {
	if _, ok := cf.Listeners[name]; ok {
		return errors.New("Listener already exists")
	} else {
		if cf.RouteOk(route) {
			cf.Listeners[name] = &Listener{name, address, port, route, StateEnabled}
			_ = cf.RouteAssign(route, name)
		} else {
			cf.Listeners[name] = &Listener{name, address, port, route, StateDisabled}
		}
		err := cf.GenerateSnapshot()
		return err
	}
}

func (cf Configuration) ListenerCheck(route string) {
	for _, l := range cf.Listeners {
		if l.State == StateDisabled {
			if _, ok := cf.RouteConf[l.Route]; ok {
				l.State = StateEnabled
				cf.RouteConf[route].Assigments[l.Name] = true
			}
		}
	}
}

func (cf Configuration) AddMirroring(route, cluster string, fraction uint32) error {
	cf.RouteConf[route].Mirroring[cluster] = fraction
	err := cf.GenerateSnapshot()
	return err
}

func (cf Configuration) GenerateSnapshot() error {
	var endpoints, clusters, routes, listeners []types.Resource
	for _, elem := range cf.Clusters {
		clusters = append(clusters, makeCluster(elem.Name))
		endpoints = append(endpoints, makeEndpoint(elem))
	}

	for _, elem := range cf.RouteConf {
		if len(elem.Assigments) > 0 && len(elem.Cluster) > 0 {
			r := makeRoute(elem.Name, elem.Cluster)
			if len(elem.Mirroring) > 0 {
				var m []*route.RouteAction_RequestMirrorPolicy
				for mirror, fraction := range elem.Mirroring {
					m = makeMirroringConfig(mirror, fraction)
				}
				r.VirtualHosts[0].Routes[0].GetRoute().RequestMirrorPolicies = m
			}
			routes = append(routes, r)
		} else {
			Log.Infof("route %s has 0 assigments, skipping", elem.Name)
		}
	}

	for _, elem := range cf.Listeners {
		if elem.State == StateEnabled {
			listeners = append(listeners, makeHTTPListener(elem))
		} else {
			Log.Infof("listener '%s' is disabled, skipping", elem.Name)
		}
	}

	cache_id := time.Now().Unix()
	snapshot := cache.NewSnapshot(
		strconv.FormatInt(cache_id, 16),
		endpoints,
		clusters,
		routes,
		listeners,
		[]types.Resource{}, // runtimes
		[]types.Resource{}, // secrets
	)

	if err := snapshot.Consistent(); err != nil {
		Log.Errorf("snapshot inconsistency: %+v\n%+v", snapshot, err)
		return err
	}

	scache := *cf.SnapshotCache
	if err := scache.SetSnapshot(nodeID, snapshot); err != nil {
		Log.Errorf("snapshot error %q for %+v", err, snapshot)
		return err
	} else {
		return nil
	}
}
