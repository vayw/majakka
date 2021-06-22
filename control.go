package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CInfo(c *gin.Context) {
	c.JSON(http.StatusOK, CF)
}

func AddListener(c *gin.Context) {
	var data ListenerRequest
	c.BindJSON(&data)
	err := CF.AddListener(data.Name, data.Address, data.Port, data.Route)
	if err != nil {
		c.JSON(http.StatusOK, err)
	} else {
		c.JSON(http.StatusCreated, "Listener created")
	}
}

func AddCluster(c *gin.Context) {
	var data ClusterRequest
	c.BindJSON(&data)
	if err := CF.AddCluster(data.Name); err != nil {
		c.JSON(http.StatusAlreadyReported, err)
	} else {
		c.JSON(http.StatusCreated, "Cluster created")
	}
}

func AddRoute(c *gin.Context) {
	var data RouteRequest
	c.BindJSON(&data)
	if err := CF.AddRoute(data.Name, data.ClusterName); err != nil {
		c.JSON(http.StatusAlreadyReported, err)
	} else {
		c.JSON(http.StatusCreated, "Route created")
	}
}

func AddEndpoint(c *gin.Context) {
	var data EndpointRequest
	c.BindJSON(&data)
	if err := CF.AddEndpoint(data.Name, data.ClusterName, data.Address, data.Port); err == nil {
		c.JSON(http.StatusCreated, "Endpoint added")
	} else {
		c.JSON(http.StatusFailedDependency, err)
	}
}
