package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CInfo(c *gin.Context) {
	c.JSON(http.StatusOK, Clusters)
}

func AddListener(c *gin.Context) {
	var data ListenerRequest
	c.BindJSON(&data)
	Listeners[data.Name] = Listener{data.Name, data.Address, data.Port, data.Route}

	if _, ok := Routes[data.Route]; ok {
		c.JSON(http.StatusCreated, "Listener created")
		err := GenerateSnapshot()
		if err != nil {
			c.JSON(http.StatusOK, err)
		}
	} else {
		c.JSON(http.StatusAccepted, "Listener created, but not started: route not found")
	}
}

func AddCluster(c *gin.Context) {
	var data ClusterRequest
	c.BindJSON(&data)
	if _, ok := Clusters[data.Name]; ok {
		c.JSON(http.StatusAlreadyReported, "Cluster already created")
	} else {
		Clusters[data.Name] = Cluster{data.Name, EndpointsMap{}}
		c.JSON(http.StatusCreated, "Cluster created")
		err := GenerateSnapshot()
		if err != nil {
			c.JSON(http.StatusOK, err)
		}
	}
}

func AddRoute(c *gin.Context) {
	var data RouteRequest
	c.BindJSON(&data)
	if _, ok := Routes[data.Name]; ok {
		c.JSON(http.StatusAlreadyReported, "Route already created")
	} else {
		Routes[data.Name] = Route{data.Name, data.ClusterName}
		c.JSON(http.StatusCreated, "Route created")
		err := GenerateSnapshot()
		if err != nil {
			c.JSON(http.StatusOK, err)
		}
	}
}

func AddEndpoint(c *gin.Context) {
	var data EndpointRequest
	c.BindJSON(&data)
	if _, ok := Clusters[data.ClusterName]; ok {
		Clusters[data.ClusterName].Endpoints[data.Name] = Endpoint{data.Host, data.Port}
		c.JSON(http.StatusCreated, "Endpoint added")
		err := GenerateSnapshot()
		if err != nil {
			c.JSON(http.StatusOK, err)
		}
	} else {
		c.JSON(http.StatusFailedDependency, "cluster not found!")
	}

}
