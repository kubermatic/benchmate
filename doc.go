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

// Package benchmate simplifies the construction of clients and servers
// for network throughput/latency estimation. This lets you easily construct your
// own tools instead of having to wrap existing tools and parse their command
// line output. For example tools see https://github.com/kubermatic/benchmate/tree/master/cmd.
//
// The package also contains HTTP handlers
// (ThroughputHandler, LatencyHandler) that can be added to your programs
// so that they can participate in network performance estimation. Like pprof [1]
// but for networking.
//
// [1] https://pkg.go.dev/net/http/pprof
package benchmate
