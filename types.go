package main

type ClustersMap map[string]Cluster
type RoutesMap map[string]Route
type ListenersMap map[string]Listener
type EndpointsMap map[string]Endpoint

type Cluster struct {
	Name      string
	Endpoints EndpointsMap
}

type Route struct {
	Name    string
	Cluster string
}

type Endpoint struct {
	UpstreamHost string
	UpstreamPort uint32
}

type Listener struct {
	Name    string
	Address string
	Port    uint32
	Route   string
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

type EndpointRequest struct {
	Name        string `json:"name" binding:"required"`
	ClusterName string `json:"cluster" binding:"required"`
	Host        string `json:"host" binding:"required"`
	Port        uint32 `json:"port" binding:"required"`
}
