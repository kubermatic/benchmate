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

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kubermatic/benchmate/pkg/latency"
	"github.com/kubermatic/benchmate/pkg/throughput"
	"google.golang.org/grpc"
	"log"
	"net"
	"sigs.k8s.io/apiserver-network-proxy/konnectivity-client/pkg/client"
)

func main() {

	var proxyUDSName string
	flag.StringVar(&proxyUDSName, "proxy-uds", "/etc/kubernetes/konnectivity-server/konnectivity-server.socket", "konnectivity-benchmate socket name of konnectivity proxy")

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

	tpResult, err := throughput.NewThroughputMeter(throughput.DefaultOptions()).ClientConn(proxyConn)
	if err != nil {
		log.Println(err)
	}

	fmt.Println(tpResult)

	requestAddress = fmt.Sprintf("%s:%d", nodeIP, 13501)
	proxyConn, err = tunnel.DialContext(ctx, "tcp", requestAddress)
	if err != nil {
		panic(err)
	}

	latResult, err := latency.NewLatencyMeter(latency.DefaultOptions()).ClientConn(proxyConn)
	if err != nil {
		log.Println(err)
	}

	fmt.Println(latResult)

}
