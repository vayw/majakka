package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func CInfo(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}

func AddListener(c *gin.Context) {
	var data ListenerRequest
	c.BindJSON(&data)
	Listeners = append(Listeners, makeHTTPListener(data.Name, data.Route))

	snapshot := GenerateSnapshot()

	if err := snapshot.Consistent(); err != nil {
		l.Errorf("snapshot inconsistency: %+v\n%+v", snapshot, err)
		os.Exit(1)
	}

	if err := SCache.SetSnapshot(nodeID, snapshot); err != nil {
		l.Errorf("snapshot error %q for %+v", err, snapshot)
		os.Exit(1)
	}
}

func AddCluster(c *gin.Context) {
	var data ClusterRequest
	c.BindJSON(&data)
	Clusters = append(Clusters, makeCluster(data.Name))

	snapshot := GenerateSnapshot()

	if err := snapshot.Consistent(); err != nil {
		l.Errorf("snapshot inconsistency: %+v\n%+v", snapshot, err)
		os.Exit(1)
	}

	if err := SCache.SetSnapshot(nodeID, snapshot); err != nil {
		l.Errorf("snapshot error %q for %+v", err, snapshot)
		os.Exit(1)
	}
}

func AddRoute(c *gin.Context) {
	var data RouteRequest
	c.BindJSON(&data)
	Routes = append(Routes, makeRoute(data.Name, data.ClusterName))

	snapshot := GenerateSnapshot()

	if err := snapshot.Consistent(); err != nil {
		l.Errorf("snapshot inconsistency: %+v\n%+v", snapshot, err)
		os.Exit(1)
	}

	if err := SCache.SetSnapshot(nodeID, snapshot); err != nil {
		l.Errorf("snapshot error %q for %+v", err, snapshot)
		os.Exit(1)
	}
}
