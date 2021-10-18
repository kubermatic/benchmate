

uds-benchmate: main.go
	CGO_ENABLED=0 go build -o uds-benchmate
	docker build -t quay.io/pratikdeoghare/uds-benchmate .

quay-push: main.go
	docker push quay.io/pratikdeoghare/uds-benchmate

clean:
	rm uds-benchmate
	docker rmi quay.io/pratikdeoghare/uds-benchmate

