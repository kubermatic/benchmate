package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kubermatic/benchmate/latency"
	"github.com/kubermatic/benchmate/throughput"
	"google.golang.org/grpc"
	"log"
	"net"
	"sigs.k8s.io/apiserver-network-proxy/konnectivity-client/pkg/client"
)

func main() {

	var proxyUDSName string
	flag.StringVar(&proxyUDSName, "proxy-uds", "/etc/kubernetes/konnectivity-server/konnectivity-server.socket", "uds-benchmate socket name of konnectivity proxy")

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

	err = throughput.ClientConn(proxyConn)
	if err != nil {
		log.Println(err)
	}

	requestAddress = fmt.Sprintf("%s:%d", nodeIP, 13501)
	proxyConn, err = tunnel.DialContext(ctx, "tcp", requestAddress)
	if err != nil {
		panic(err)
	}

	err = latency.ClientConn(proxyConn)
	if err != nil {
		log.Println(err)
	}

}
