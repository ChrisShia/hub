up_build: build_hub
	@echo "Stopping docker images (if running...)"
	docker-compose down
	@echo "Starting docker images..."
	docker-compose up --build -d
	@echo "Docker images started!"

build_hub:
	@echo "Building hub-service binary..."
	cd hub-service && env GOOS=linux go build -o ./build/hubApp ./cmd/api
	@echo "Done!"

hub_down:
	@echo "Stopping hub-service Docker image..."
	docker-compose down hub-service
	@echo "Done"
	
hub: hub_down build_hub 
	@echo "Starting new hub-service docker image..."
	docker-compose up --build -d hub-service
	@echo "Docker image started!"

down:
	@echo "Stopping docker images..."
	docker-compose down
	@echo "Done!"

up:
	@echo "Starting Docker images..."
	docker-compose up -d
	@echo "Docker images started!"
