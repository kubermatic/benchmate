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

package benchmate

// Options contains configuration options for clients and servers.
type Options struct {
	MsgSize    int    `json:"msgSize"`    // size of messages in bytes
	NumMsg     int    `json:"numMsg"`     // number of messages to send
	Addr       string `json:"addr"`       // server listens on this address
	Network    string `json:"network"`    // network type (unix or tcp)
	ClientPort int    `json:"clientPort"` // local port used by client
	Timeout    int    `json:"timeout"`    // in milliseconds
}

// LatencyServer returns a LatencyServer configured with the options.
func (o Options) LatencyServer() LatencyServer {
	return LatencyServer{
		msgSize: o.MsgSize,
		numMsg:  o.NumMsg,
	}
}

// LatencyClient returns a LatencyClient configured with the options.
func (o Options) LatencyClient() LatencyClient {
	return LatencyClient{
		msgSize: o.MsgSize,
		numMsg:  o.NumMsg,
		timeout: o.Timeout,
	}
}

// ThroughputServer returns a ThroughputServer configured with the options.
func (o Options) ThroughputServer() ThroughputServer {
	return ThroughputServer{
		msgSize: o.MsgSize,
	}
}

// ThroughputClient returns a ThroughputClient configured with the options.
func (o Options) ThroughputClient() ThroughputClient {
	return ThroughputClient{
		msgSize: o.MsgSize,
		numMsg:  o.NumMsg,
		timeout: o.Timeout,
	}
}

// DefaultLatencyOptions
func DefaultLatencyOptions() Options {
	return Options{
		MsgSize:    128,
		NumMsg:     10000,
		Addr:       ":13501",
		Network:    "tcp",
		ClientPort: 0,
		Timeout:    120000,
	}
}

// DefaultThroughputOptions
func DefaultThroughputOptions() Options {
	return Options{
		MsgSize:    256 * 1024,
		NumMsg:     10000,
		Addr:       ":13500",
		Network:    "tcp",
		ClientPort: 0,
		Timeout:    120000,
	}
}
