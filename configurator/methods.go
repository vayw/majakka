import (
	"strconv"
	"time"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v2"
)

func (cf *Configuration) GenerateSnapshot() error {
	var endpoints, clusters, routes, listeners []types.Resource
	for _, elem := range cf.Clusters {
		clusters = append(clusters, makeCluster(elem.Name))
		endpoints = append(endpoints, makeEndpoint(elem))
	}

	for _, elem := range cf.RouteConf {
		if len(elem.Assigments) > 0 && len(elem.VHosts) > 0 {
			r := makeRoute(elem.Name)
			routes = append(routes, r)
		} else {
			Log.Infof("route %s has 0 assigments, skipping", elem.Name)
		}
	}

	for _, elem := range cf.Listeners {
		if elem.State == StateEnabled {
			listeners = append(listeners, elem.Generate())
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
