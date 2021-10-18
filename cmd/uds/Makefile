

benchmate: main.go
	CGO_ENABLED=0 go build 
	docker build -t quay.io/pratikdeoghare/benchmate .

quay-push: main.go
	docker push quay.io/pratikdeoghare/benchmate

run-tcp-server:
	docker run -it --rm --net=host benchmate

run-tcp-client:
	docker run -it --rm --net=host benchmate -c

run-uds-server:
	docker run -it --rm --net=host -v /tmp/lat_benchmark.sock:/tmp/lat_benchmark.sock -v /tmp/tp_benchmark.sock:/tmp/tp_benchmark.sock benchmate benchmate -unixdomain

run-uds-client:
	docker run -it --rm --net=host -v /tmp/lat_benchmark.sock:/tmp/lat_benchmark.sock -v /tmp/tp_benchmark.sock:/tmp/tp_benchmark.sock benchmate -c -unixdomain

clean:
	rm benchmate 
	docker rmi quay.io/pratikdeoghare/benchmate
