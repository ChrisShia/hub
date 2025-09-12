.PHONY: help 
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## up_build: Stop all docker images (if running), start them again, building them if necessary and start them up.
.PHONY: up/build
up/build: hub/build
	@echo "Stopping docker images (if running...)"
	docker-compose down
	@echo "Starting docker images..."
	docker-compose up --build -d
	@echo "Docker images started!"

## down: Stop all docker images.
.PHONY: down
down:
	@echo "Stopping docker images..."
	docker-compose down
	@echo "Done!"

## up: Start all docker images.
.PHONY: up
up:
	@echo "Starting Docker images..."
	docker-compose up -d
	@echo "Docker images started!"

## hub/build/up: Execute hub/down, hub/build, build hub image again and start container.
.PHONY: hub/build/up
hub/build/up: hub/down hub/build 
	@echo "Starting new hub-service docker image..."
	docker-compose up --build -d hub-service
	@echo "Docker image started!"

## hub/build: Build hub-service binary.
.PHONY: hub/build
hub/build:
	@echo "Building hub-service binary..."
	cd hub-service && env GOOS=linux go build -o ./build/hubApp ./cmd/api
	@echo "Done!"

## hub/down: Stop hub-service Docker image.
.PHONY: hub/down
hub/down:
	@echo "Stopping hub-service Docker image..."
	docker-compose down hub-service
	@echo "Done"

## hub/up: Start hub-service docker container, with existing image (if any).
.PHONY: hub/up
hub/up:
	@echo "Starting hub-service docker image"
	docker-compose up -d hub-service
	@echo "Docker image started!"
