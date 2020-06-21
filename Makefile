
TAG=sudachen/local-testnet-app

build:
	go build -o .bin/local-testnet .
	docker build -t sudachen/local-testnet-app .
push:
	docker push sudachen/local-testnet-app


