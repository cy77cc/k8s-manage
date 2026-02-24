swag:
	swag init

run:
	go run main.go

web-build:
	cd web && npm run build

build:
	go build -o bin/k8s-manage main.go

build-all: web-build build

docker:
	docker buildx build -t k8s-manage .

migrate-up:
	go run main.go migrate up

migrate-status:
	go run main.go migrate status

migrate-down:
	go run main.go migrate down --steps=1
