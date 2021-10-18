package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/pratikdeoghare/benchmate/latency"
	"github.com/pratikdeoghare/benchmate/throughput"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
	"log"
	"net"
	"sigs.k8s.io/apiserver-network-proxy/konnectivity-client/pkg/client"
)

func main() {

	var proxyUDSName string
	flag.StringVar(&proxyUDSName, "proxy-uds", "/etc/kubernetes/konnectivity-server/konnectivity-server.socket", "uds socket name of konnectivity proxy")

	var nodeIP string
	flag.StringVar(&nodeIP, "node-ip", "127.0.0.1", "ip of node where benchmate server is running")

	flag.Parse()

	dialOption := grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		// Ignoring addr and timeout arguments:
		// addr - comes from the closure
		// timeout - is turned off as this is test code and eases debugging.
		c, err := net.DialTimeout("unix", proxyUDSName, 0)
		if err != nil {
			klog.ErrorS(err, "failed to create connection to uds", "name", proxyUDSName)
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

	//
	//for {
	//	n, err := proxyConn.Write([]byte("Hello there"))
	//	if err != nil {
	//		panic(err)
	//	}
	//	fmt.Println("wrote ",n, " bytes")
	//	time.Sleep(time.Second)
	//}

}
