// Copyright 2020 Envoyproxy Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/envoyproxy/go-control-plane/pkg/test/v3"
	"github.com/gin-gonic/gin"
)

var (
	Log Logger

	port     uint
	basePort uint
	mode     string

	nodeID string

	CF Configuration

	SCache cache.SnapshotCache
)

func init() {
	Log = Logger{}

	flag.BoolVar(&Log.Debug, "debug", false, "Enable xDS server debug logging")

	// The port that this xDS server listens on
	flag.UintVar(&port, "port", 18000, "xDS management server port")

	// Tell Envoy to use this Node ID
	flag.StringVar(&nodeID, "nodeID", "test-id", "Node ID")
}

func main() {
	flag.Parse()

	CF = Configuration{
		Clusters:  make(ClustersMap),
		Listeners: make(ListenersMap),
		RouteConf: make(RouteConfMap),
	}

	controlapi := gin.Default()
	controlapi.GET("/control/info", CInfo)
	controlapi.POST("/control/listener/add", AddListener)
	controlapi.POST("/control/cluster/add", AddCluster)
	controlapi.POST("/control/route/add", AddRouteConf)
	controlapi.POST("/control/endpoint/add", AddEndpoint)
	controlapi.POST("/control/endpoint/delete", DeleteEndpoint)
	controlapi.POST("/control/endpoint/switch", SwitchEndpoint)
	controlapi.POST("/control/mirroring/add", AddMirroring)

	httpport := fmt.Sprintf(":8099")
	go controlapi.Run(httpport)

	// Create a cache
	SCache = cache.NewSnapshotCache(false, cache.IDHash{}, Log)
	CF.SnapshotCache = &SCache

	// Run the xDS server
	ctx := context.Background()
	cb := &test.Callbacks{Debug: Log.Debug}
	srv := server.NewServer(ctx, SCache, cb)
	RunServer(ctx, srv, port)
}
