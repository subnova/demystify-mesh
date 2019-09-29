package main

import (
	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/fsnotify/fsnotify"
	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/grpc"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"net"
	"path/filepath"
	"strconv"
)

var (
	address = kingpin.Flag("address", "The address to bind to").Default(":8080").String()
	clusters = kingpin.Arg("clusters", "The file containing the cluster configuration").Required().String()
	routes = kingpin.Arg("routes", "The file containing the route configuration").Required().String()
)

type DummyNodeHash struct{}

func (*DummyNodeHash) ID(node *core.Node) string {
	return "dummy"
}

func readClusters(data string) (*map[string]cache.Resource, error) {
	json, err := yaml.YAMLToJSON([]byte(data))
	if err != nil {
		return nil, err
	}
	clusters := Clusters{}
	err = jsonpb.UnmarshalString(string(json), &clusters)
	if err != nil {
		return nil, err
	}

	result := make(map[string]cache.Resource, len(clusters.Clusters))

	for _, cluster := range clusters.Clusters {
		result[cache.GetResourceName(cluster)] = cluster
	}

	return &result, nil
}

func readRoutes(data string) (*map[string]cache.Resource, error) {
	json, err := yaml.YAMLToJSON([]byte(data))
	if err != nil {
		return nil, err
	}
	routes := Routes{}
	err = jsonpb.UnmarshalString(string(json), &routes)
	if err != nil {
		return nil, err
	}

	result := make(map[string]cache.Resource, len(routes.Routes))

	for _, route := range routes.Routes {
		result[cache.GetResourceName(route)] = route
	}

	return &result, nil
}

func main() {
	kingpin.Parse()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	snapshotCache := cache.NewSnapshotCache(false, &DummyNodeHash{}, nil)

	version := 1

	go func() {
		for {
			select {
			case _, ok := <-watcher.Events:
				if !ok {
					return
				}

				err := updateCache(snapshotCache, version)
				if err != nil {
					panic(err)
				}

				version++
			}
		}
	}()

	_ = watcher.Add(filepath.Dir(*clusters))
	_ = watcher.Add(filepath.Dir(*routes))

	err = updateCache(snapshotCache, 0)
	if err != nil {
		panic(err)
	}

	newServer := server.NewServer(snapshotCache, nil)
	var opts []grpc.ServerOption
	s := grpc.NewServer(opts...)
	envoy.RegisterClusterDiscoveryServiceServer(s, newServer)
	envoy.RegisterRouteDiscoveryServiceServer(s, newServer)
	listener, err := net.Listen("tcp", *address)
	if err != nil {
		panic(err)
	}
	err = s.Serve(listener)
	if err != nil {
		panic(err)
	}
}

func updateCache(snapshotCache cache.SnapshotCache, version int) error {
	data, err := ioutil.ReadFile(*clusters)
	if err != nil {
		return err
	}
	cdsClusters, err := readClusters(string(data))
	if err != nil {
		return err
	}
	data, err = ioutil.ReadFile(*routes)
	if err != nil {
		return err
	}
	rdsRoutes, err := readRoutes(string(data))
	if err != nil {
		return err
	}
	err = snapshotCache.SetSnapshot("dummy", cache.Snapshot{
		Clusters: cache.Resources{
			Version: strconv.Itoa(version),
			Items:   *cdsClusters,
		},
		Routes: cache.Resources{
			Version: strconv.Itoa(version),
			Items:   *rdsRoutes,
		},
	})
	if err != nil {
		return err
	}

	return nil
}
