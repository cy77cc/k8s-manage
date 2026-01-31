swag:
	swag init

run:
	go run main.go

build:
	go build -o bin/k8s-mange main.go

docker:
	docker buildx build -t k8s-manage .