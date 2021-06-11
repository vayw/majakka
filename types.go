package main

import "github.com/envoyproxy/go-control-plane/pkg/cache/types"

type ClusterList []types.Resource
type RouteList []types.Resource
type EndpointList []types.Resource
type ListenerList []types.Resource

type ListenerRequest struct {
	Name  string `json:"name" binding:"required"`
	Route string `json:"route" binding:"required"`
}

type ClusterRequest struct {
	Name string `json:"name" binding:"required"`
}

type RouteRequest struct {
	Name        string `json:"name" binding:"required"`
	ClusterName string `json:"cluster" binding:"required"`
	//UpstreamHost string `json:"host" binding:"required"`
	//UpstreamPort int    `json:"port" binding:"required"`
}
