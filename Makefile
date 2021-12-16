#	Copyright 2021 The Kubermatic Kubernetes Platform contributors.
#
#	Licensed under the Apache License, Version 2.0 (the "License");
#	you may not use this file except in compliance with the License.
#	You may obtain a copy of the License at
#
#		http://www.apache.org/licenses/LICENSE-2.0
#
#	Unless required by applicable law or agreed to in writing, software
#	distributed under the License is distributed on an "AS IS" BASIS,
#	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#	See the License for the specific language governing permissions and
#	limitations under the License.
#

benchmate: pkg cmd/benchmate
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o _build/benchmate cmd/benchmate/main.go

docker-build-benchmate: benchmate
	docker build -t quay.io/kubermatic-labs/benchmate:latest -f benchmate.Dockerfile .

konnectivity-benchmate: pkg cmd/konnectivity-benchmate
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o _build/konnectivity-benchmate cmd/konnectivity-benchmate/main.go

docker-build-konnectivity-benchmate: konnectivity-benchmate
	docker build -t quay.io/kubermatic-labs/konnectivity-benchmate:latest -f konnectivity-benchmate.Dockerfile .

lint:
	golangci-lint run \
		--verbose \
		--print-resources-usage \
		./...
build:
	go build -v ./...

test:
	go test -v -race ./...

clean:
	rm -rf _build
