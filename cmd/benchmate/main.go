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

// Using benchmate you can estimate latency and throughput of a network.
//
// You run the server somewhere like
//	 $ benchmate
// You run the client somewhere like
// 	$ benchmate -c
//
// As long as the client can talk to the server, you will get estimates at the client.
//
//	Usage of ./benchmate:
//		-addr string
//			set the address (default ":12345")
//		-c	set the flag to run in client mode. Default is server mode.
//		-clientPort int
//			set the client port (valid only in client mode)
//		-lat
//			set the flag to run in latency mode and specify the options on command line
//		-latOpt string
//			set the latency options using json file
//		-msgSize int
//			set the message size (default 1024)
//		-network string
//			set the network (tcp or unix) (default "tcp")
//		-numMsg int
//			set the number of messages to exchange (default 1000)
//		-timeout int
//			set the timeout (ms) (default 120000)
//		-tp
//			set the flag to run in throughput mode and specify the options on command line
//		-tpOpt string
//			set the throughput options using json file
//
// You can specify options using a json files using --tpOpt, --latOpt parameters.
// Valid format of the json files is here http://pkg.go.dev/github.com/kubermatic/benchmate/#Options
//
package main

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"time"

	"github.com/kubermatic/benchmate"
)

func main() {
	log.SetFlags(0)
	var c bool

	var latOptFile string
	var tpOptFile string

	var lat bool
	var tp bool

	var (
		msgSize    int
		numMsg     int
		addr       string
		network    string
		clientPort int
		timeout    int
	)

	flag.BoolVar(&c, "c", false, "set the flag to run in client mode. Default is server mode. ")

	flag.StringVar(&latOptFile, "latOpt", "", "set the latency options using json file")
	flag.StringVar(&tpOptFile, "tpOpt", "", "set the throughput options using json file")

	flag.BoolVar(&lat, "lat", false, "set the flag to run in latency mode and specify the options on command line")
	flag.BoolVar(&tp, "tp", false, "set the flag to run in throughput mode and specify the options on command line")

	flag.IntVar(&msgSize, "msgSize", 1024, "set the message size")
	flag.IntVar(&numMsg, "numMsg", 1000, "set the number of messages to exchange")
	flag.StringVar(&addr, "addr", ":12345", "set the address")
	flag.StringVar(&network, "network", "tcp", "set the network (tcp or unix)")
	flag.IntVar(&clientPort, "clientPort", 0, "set the client port (valid only in client mode)")
	flag.IntVar(&timeout, "timeout", 120000, "set the timeout (ms)")

	flag.Parse()

	var latOpts benchmate.Options
	if latOptFile != "" {
		latOpts = benchmate.DefaultLatencyOptions()
		data, err := ioutil.ReadFile(latOptFile)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(data, &latOpts)
		if err != nil {
			log.Fatal(err)
		}
	}

	var tpOpts benchmate.Options
	if tpOptFile != "" {
		tpOpts = benchmate.DefaultThroughputOptions()
		data, err := ioutil.ReadFile(tpOptFile)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(data, &tpOpts)
		if err != nil {
			log.Fatal(err)
		}
	}

	if c {
		if latOptFile != "" {
			runLatencyClient(latOpts)
		}
		if tpOptFile != "" {
			runThroughputClient(tpOpts)
		}
	} else {
		var wg sync.WaitGroup
		if latOptFile != "" {
			wg.Add(1)
			go func() {
				defer wg.Done()
				runLatencyServer(latOpts)
			}()
		}
		if tpOptFile != "" {
			wg.Add(1)
			go func() {
				defer wg.Done()
				runThroughputServer(tpOpts)
			}()
		}
		wg.Wait()
	}

	// If options are specified using json files then ignore the command line options.
	if latOptFile != "" || tpOptFile != "" {
		return
	}

	if lat && tp {
		log.Fatal("cannot run both latency and throughput with command line flags provide options with JSON files using --latOpts, --tpOpts flags instead.")
	}

	// run both benchmarks with defualt options when nothing is specified
	if !lat && !tp {
		latOpts := benchmate.DefaultLatencyOptions()
		tpOpts := benchmate.DefaultThroughputOptions()
		if c {
			runLatencyClient(latOpts)
			runThroughputClient(tpOpts)
		} else {
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				runLatencyServer(latOpts)
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				runThroughputServer(tpOpts)
			}()
			wg.Wait()
		}
		return
	}

	var opts benchmate.Options
	if lat {
		opts = benchmate.DefaultLatencyOptions()
	} else if tp {
		opts = benchmate.DefaultThroughputOptions()
	}

	// override the options with command line flags
	if isFlagPassed("msgSize") {
		opts.MsgSize = msgSize
	}
	if isFlagPassed("numMsg") {
		opts.NumMsg = numMsg
	}
	if isFlagPassed("addr") {
		opts.Addr = addr
	}
	if isFlagPassed("network") {
		opts.Network = network
	}
	if isFlagPassed("clientPort") {
		opts.ClientPort = clientPort
	}
	if isFlagPassed("timeout") {
		opts.Timeout = timeout
	}

	if lat {
		if c {
			runLatencyClient(opts)
		} else {
			runLatencyServer(opts)
		}
	} else if tp {
		if c {
			runThroughputClient(opts)
		} else {
			runThroughputServer(opts)
		}
	}

	log.Println("done.")
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func prettyJSON(x interface{}) string {
	b, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}

func runThroughputClient(tpOpt benchmate.Options) {
	log.Println("running throughput client with:", prettyJSON(tpOpt))
	conn, err := net.Dial(tpOpt.Network, tpOpt.Addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	tpResult, err := tpOpt.ThroughputClient().Run(conn)
	if err != nil {
		log.Println("throughput measurement failed:", err)
	} else {
		log.Println("throughput benchmark result:", prettyJSON(tpResult))
		log.Println("throughput: ", float64(tpResult.NumMsg*tpResult.MsgSize*1000)/float64(tpResult.Elapsed.Nanoseconds()), "MB/s")
		log.Println("throughput client done.")
	}
}

func runLatencyClient(latOpt benchmate.Options) {
	log.Println("running latency client with:", prettyJSON(latOpt))
	conn, err := net.Dial(latOpt.Network, latOpt.Addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	latResult, err := latOpt.LatencyClient().Run(conn)
	if err != nil {
		log.Println("latency measurement failed:", err)
	} else {
		log.Println("latency benchmark result:", prettyJSON(latResult))
		log.Println("average latency:", time.Duration(float64(latResult.ElapsedTime.Nanoseconds())/float64(latResult.NumMsg)))
		log.Println("latency client done.")
	}
}

func runThroughputServer(tpOpt benchmate.Options) {

	l, err := net.Listen(tpOpt.Network, tpOpt.Addr)
	if err != nil {
		log.Println("throughput server failed:", err)
		return
	}
	defer l.Close()

	log.Println("running throughput server with:", prettyJSON(tpOpt))

	err = tpOpt.ThroughputServer().Run(l)
	if err != nil {
		if err == io.EOF {
			log.Println("throughput server done.")
		} else {
			log.Println("throughput server failed:", err)
		}
	}

}

func runLatencyServer(latOpt benchmate.Options) {
	l, err := net.Listen(latOpt.Network, latOpt.Addr)
	if err != nil {
		log.Println("latency server failed:", err)
		return
	}
	defer l.Close()

	log.Println("running latency server with:", prettyJSON(latOpt))

	err = latOpt.LatencyServer().Run(l)
	if err != nil {
		log.Println("latency server:", err)
	} else {
		log.Println("latency server done.")
	}
}
