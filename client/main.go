package main

import (
	"context"
	"flag"
	xdsapi "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"google.golang.org/grpc"
	"log"
)

var (
	endpoint = flag.String("endpoint", "127.0.0.1:8080", "xDS endpoint")
)

func main() {
	flag.Parse()
	conn, err := grpc.Dial(*endpoint, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return
	}
	ctx := context.Background()

	cdsclient := xdsapi.NewClusterDiscoveryServiceClient(conn)

	resp, err := cdsclient.FetchClusters(ctx, &xdsapi.DiscoveryRequest{
		TypeUrl: "type.googleapis.com/envoy.api.v2.Cluster",
		Node:    &core.Node{},
	})
	if err != nil {
		log.Fatalf("get Clusters resp error: %v", err)
		return
	}
	clusters := make([]*xdsapi.Cluster, 0)
	for _, res := range resp.Resources {
		cluster := xdsapi.Cluster{}
		_ = cluster.XXX_Unmarshal(res.GetValue())
		clusters = append(clusters, &cluster)
	}

	log.Println("-----Clusters-------")
	for _, cluster := range clusters {
		log.Printf("%v", cluster)
	}

	rdsclient := xdsapi.NewRouteDiscoveryServiceClient(conn)

	resp, err = rdsclient.FetchRoutes(ctx, &xdsapi.DiscoveryRequest{
		TypeUrl: "type.googleapis.com/envoy.api.v2.RouteConfiguration",
		Node:    &core.Node{},
	})
	if err != nil {
		log.Fatalf("get Routes resp error: %v", err)
		return
	}
	routes := make([]*xdsapi.RouteConfiguration, 0)
	for _, res := range resp.Resources {
		route := xdsapi.RouteConfiguration{}
		_ = route.XXX_Unmarshal(res.GetValue())
		routes = append(routes, &route)
	}
	log.Println("-----routes-------")
	for _, route := range routes {
		log.Printf("%v", route)
	}

}
