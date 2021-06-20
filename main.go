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
	l Logger

	port     uint
	basePort uint
	mode     string

	nodeID string

	Listeners ListenersMap
	Clusters  ClustersMap
	Routes    RoutesMap

	SCache cache.SnapshotCache
)

func init() {
	l = Logger{}

	flag.BoolVar(&l.Debug, "debug", false, "Enable xDS server debug logging")

	// The port that this xDS server listens on
	flag.UintVar(&port, "port", 18000, "xDS management server port")

	// Tell Envoy to use this Node ID
	flag.StringVar(&nodeID, "nodeID", "test-id", "Node ID")
}

func main() {
	flag.Parse()

	Listeners = make(ListenersMap)
	Clusters = make(ClustersMap)
	Routes = make(RoutesMap)

	controlapi := gin.Default()
	controlapi.GET("/control/info", CInfo)
	controlapi.POST("/control/listener/add", AddListener)
	controlapi.POST("/control/cluster/add", AddCluster)
	controlapi.POST("/control/route/add", AddRoute)
	controlapi.POST("/control/endpoint/add", AddEndpoint)

	httpport := fmt.Sprintf(":8099")
	go controlapi.Run(httpport)

	// Create a cache
	SCache = cache.NewSnapshotCache(false, cache.IDHash{}, l)

	// Run the xDS server
	ctx := context.Background()
	cb := &test.Callbacks{Debug: l.Debug}
	srv := server.NewServer(ctx, SCache, cb)
	RunServer(ctx, srv, port)
}
