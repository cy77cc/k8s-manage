swag:
	swag init

run:
	go run main.go

web-build:
	cd web && npm run build

build:
	go build -o bin/k8s-mange main.go

build-all: web-build build

docker:
	docker buildx build -t k8s-manage .
