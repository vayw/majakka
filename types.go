package main

import (
	"errors"
	"strconv"
	"time"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
)

const StateEnabled = "enabled"
const StateDisabled = "disabled"

type ClustersMap map[string]Cluster
type RoutesMap map[string]Route
type ListenersMap map[string]Listener
type EndpointsMap map[string]Endpoint
type RouteAssigments map[string]bool

type Listener struct {
	Name    string
	Address string
	Port    uint32
	Route   string
	State   string
}

type Route struct {
	Name       string
	Cluster    string
	Assigments RouteAssigments
}

type Endpoint struct {
	UpstreamHost string
	UpstreamPort uint32
}

type Cluster struct {
	Name      string
	Endpoints EndpointsMap
}

type Configuration struct {
	Clusters      ClustersMap
	Routes        RoutesMap
	Listeners     ListenersMap
	SnapshotCache *cache.SnapshotCache
}

func (cf Configuration) AddCluster(name string) error {
	if _, ok := cf.Clusters[name]; ok {
		return errors.New("Cluster already exists")
	} else {
		cf.Clusters[name] = Cluster{name, make(EndpointsMap)}
		return nil
	}
}

func (cf Configuration) AddEndpoint(name, cluster, address string, port uint32) error {
	if _, ok := cf.Clusters[cluster]; ok {
		cf.Clusters[cluster].Endpoints[name] = Endpoint{address, port}
		return nil
	} else {
		_ = cf.AddCluster(cluster)
		if err := cf.AddEndpoint(name, cluster, address, port); err != nil {
			return err
		} else {
			return nil
		}
	}
}

func (cf Configuration) AddRoute(name, cluster string) error {
	if _, ok := cf.Routes[name]; ok {
		return errors.New("Route already exists")
	} else {
		cf.Routes[name] = Route{name, cluster, make(RouteAssigments)}
		cf.ListenerCheck(name)
		err := cf.GenerateSnapshot()

		return err
	}
}

func (cf Configuration) RouteOk(name string) bool {
	if _, ok := cf.Routes[name]; ok {
		return true
	} else {
		return false
	}
}

func (cf Configuration) RouteAssign(route, listener string) error {
	cf.Routes[route].Assigments[listener] = true
	return nil
}

func (cf Configuration) AddListener(name, address string, port uint32, route string) error {
	if _, ok := cf.Listeners[name]; ok {
		return errors.New("Listener already exists")
	} else {
		if cf.RouteOk(route) {
			cf.Listeners[name] = Listener{name, address, port, route, StateEnabled}
			_ = cf.RouteAssign(route, name)
		} else {
			cf.Listeners[name] = Listener{name, address, port, route, StateDisabled}
		}
		return nil
	}
}

func (cf Configuration) ListenerCheck(route string) {
	for _, l := range cf.Listeners {
		if l.State == StateDisabled {
			if _, ok := cf.Routes[l.Route]; ok {
				l.State = StateEnabled
				cf.Routes[route].Assigments[l.Name] = true
			}
		}
	}
}

func (cf Configuration) GenerateSnapshot() error {
	cache_id := time.Now().Unix()

	var endpoints, clusters, routes, listeners []types.Resource
	for _, elem := range cf.Clusters {
		clusters = append(clusters, makeCluster(elem.Name))
		endpoints = append(endpoints, makeEndpoint(&elem))
	}

	for _, elem := range cf.Routes {
		if len(elem.Assigments) > 0 {
			routes = append(routes, makeRoute(elem.Name, elem.Cluster))
		} else {
			Log.Infof("route %s has 0 assigments, skipping", elem.Name)
		}
	}

	for _, elem := range cf.Listeners {
		if elem.State == StateEnabled {
			listeners = append(listeners, makeHTTPListener(&elem))
		} else {
			Log.Infof("listener '%s' is disabled, skipping", elem.Name)
		}
	}

	snapshot := cache.NewSnapshot(
		strconv.FormatInt(cache_id, 16),
		endpoints, // endpoints
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

type EndpointRequest struct {
	Name        string `json:"name" binding:"required"`
	ClusterName string `json:"cluster" binding:"required"`
	Address     string `json:"address" binding:"required"`
	Port        uint32 `json:"port" binding:"required"`
}

type ListenerRequest struct {
	Name    string `json:"name" binding:"required"`
	Route   string `json:"route" binding:"required"`
	Address string `json:"address" binding:"required"`
	Port    uint32 `json:"port" binding:"required"`
}

type ClusterRequest struct {
	Name string `json:"name" binding:"required"`
}

type RouteRequest struct {
	Name        string `json:"name" binding:"required"`
	ClusterName string `json:"cluster" binding:"required"`
}
