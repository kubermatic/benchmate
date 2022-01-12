/*
Copyright 2021 The Kubermatic Kubernetes Platform contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This is only the client side of benchmate that work with konnectivity-proxy.
// You can run it in the same pod as konnectivity-proxy with -proxy-uds
// set to unix domain socket that konnectivity-proxy is listening on
// and -node-ip set to ip addr of benchmate server. This benchmate
// server should be reachable by some konnectivity-agent.
//
//  ./konnectivity-benchmate -node-ip=<server ip> -proxy-uds=/tmp/uds-socket
//
// Options:
//	$ ./konnectivity-benchmate -h
//	Usage of ./konnectivity-benchmate:
//	-node-ip string
//		ip of node where benchmate server is running (default "127.0.0.1")
//	-proxy-uds string
//		uds socket of konnectivity-proxy (default "/etc/kubernetes/konnectivity-server/konnectivity-server.socket")
package main

import (
	"context"
	"flag"
	"fmt"
	"net"

	"github.com/kubermatic/benchmate"
	"google.golang.org/grpc"
	"sigs.k8s.io/apiserver-network-proxy/konnectivity-client/pkg/client"
)

func main() {
	var proxyUDSName string
	flag.StringVar(&proxyUDSName, "proxy-uds", "/etc/kubernetes/konnectivity-server/konnectivity-server.socket", "uds socket of konnectivity-proxy")

	var nodeIP string
	flag.StringVar(&nodeIP, "node-ip", "127.0.0.1", "ip of node where benchmate server is running")

	flag.Parse()

	dialOption := grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		c, err := net.DialTimeout("unix", proxyUDSName, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to create connection to %s: %+v", proxyUDSName, err)
		}
		return c, err
	})

	ctx := context.Background()
	tunnel, err := client.CreateSingleUseGrpcTunnel(ctx, proxyUDSName, dialOption, grpc.WithInsecure(), grpc.WithUserAgent("o.userAgent"))
	if err != nil {
		panic(err)
	}

	requestAddress := fmt.Sprintf("%s:%d", nodeIP, 13500)
	proxyConn, err := tunnel.DialContext(ctx, "tcp", requestAddress)
	if err != nil {
		panic(err)
	}

	tpResult, err := benchmate.DefaultThroughputOptions().ThroughputClient().Run(proxyConn)
	if err != nil {
		panic(err)
	}

	fmt.Println(tpResult)

	requestAddress = fmt.Sprintf("%s:%d", nodeIP, 13501)
	proxyConn, err = tunnel.DialContext(ctx, "tcp", requestAddress)
	if err != nil {
		panic(err)
	}

	latResult, err := benchmate.DefaultLatencyOptions().LatencyClient().Run(proxyConn)
	if err != nil {
		panic(err)
	}

	fmt.Println(latResult)
}
