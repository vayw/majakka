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

func AddRouteConf(c *gin.Context) {
	var data RouteRequest
	c.BindJSON(&data)
	if err := CF.AddRouteConf(data.Name); err != nil {
		c.JSON(http.StatusAlreadyReported, err)
	} else {
		c.JSON(http.StatusCreated, "Route created")
	}
}

func AddEndpoint(c *gin.Context) {
	var data EndpointRequest
	err := c.BindJSON(&data)
	if err != nil {
		Log.Errorf("%s", err)
	}
	if err := CF.AddEndpoint(data.Name, data.ClusterName, data.Address, data.Port); err == nil {
		c.JSON(http.StatusCreated, "Endpoint added")
	} else {
		c.JSON(http.StatusFailedDependency, err)
	}
}

func DeleteEndpoint(c *gin.Context) {
	var data EndpointRequest
	c.BindJSON(&data)
	if err := CF.DeleteEndpoint(data.Name, data.ClusterName); err == nil {
		c.JSON(http.StatusCreated, "Endpoint deleted")
	} else {
		c.JSON(http.StatusFailedDependency, err)
	}
}

func SwitchEndpoint(c *gin.Context) {
	var data EndpointRequest
	c.BindJSON(&data)
	switch data.Switch {
	case "off":
		err := CF.DisableEndpoint(data.Name, data.ClusterName)
		if err != nil {
			c.JSON(http.StatusOK, err)
		} else {
			c.JSON(http.StatusOK, "Endpoint disabled")
		}
	case "on":
		err := CF.EnableEndpoint(data.Name, data.ClusterName)
		if err != nil {
			c.JSON(http.StatusOK, err)
		} else {
			c.JSON(http.StatusOK, "Endpoint enabled")
		}
	default:
		c.JSON(http.StatusOK, "action not supported, use on/off")
	}
}

func AddMirroring(c *gin.Context) {
	var data MirrorRequest
	c.BindJSON(&data)
	err := CF.AddMirroring(data.RouteConf, data.VHost, data.Route, data.Cluster, data.Fraction)
	if err != nil {
		c.JSON(http.StatusOK, err)
	} else {
		c.JSON(http.StatusOK, "mirroring enabled")
	}
}

type MirrorRequest struct {
	RouteConf string `json:"routeconf" binding:"required"`
	VHost     string `json:"vhost" binding:"required"`
	Route     string `json:"route" binding:"required"`
	Cluster   string `json:"cluster" binding:"required"`
	Fraction  uint32 `json:"fraction" binding:"required"`
}

type EndpointRequest struct {
	Name        string `json:"name" binding:"required"`
	ClusterName string `json:"cluster" binding:"required"`
	Address     string `json:"address"`
	Port        uint32 `json:"port"`
	Switch      string `json:"switch"`
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
	Name string `json:"name" binding:"required"`
}

type VHostRequest struct {
	Name        string      `json:"name" binding:"required"`
	RouteConf   string      `json:"routeconf" binding:"required"`
	Domains     []string    `json:"domains" binding:"required"`
	TLSOnly     bool        `json:"tlsonly"`
	TLSCertPath string      `json:"tlscertpath"`
	TLSKeyPath  string      `json:"tlskeypath"`
	Routes      VHostRoutes `json:"cluster" binding:"required"`
}
