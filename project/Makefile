AUTH_BINARY=authApp
USER_BINARY=userApp
MEDIA_BINARY=mediaApp
## up: starts all containers in the background without forcing build
up:
	@echo "Starting Docker images..."
	docker-compose up -d
	@echo "Docker images started!"

## up_build: stops docker-compose (if running), builds all projects and starts docker compose
up_build: build_auth build_user build_media build_image_proccessing_service
	@echo "Stopping docker images (if running...)"
	docker-compose down 
	@echo "Building (when required) and starting docker images..."
	docker-compose up --build
	@echo "Docker images built and started!"

## down: stop docker compose
down:
	@echo "Stopping docker compose..."
	docker-compose down
	@echo "Done!"

## build_auth: builds the auth binary as a linux executable
build_auth:
	@echo "Building auth binary..."
	cd ../authentication_service && env GOOS=linux CGO_ENABLED=0 go build -o ${AUTH_BINARY} ./cmd/api
	@echo "Done!"

## build_user: builds the users binary as a linux executable
build_user:
	@echo "Building user binary..."
	cd ../user_service && env GOOS=linux CGO_ENABLED=0 go build -o ${USER_BINARY} ./cmd/api
	@echo "Done!"

## build_user: builds the media binary as a linux executable
build_media:
	@echo "Building media binary..."
	cd ../media_service && env GOOS=linux CGO_ENABLED=0 go build -o ${MEDIA_BINARY} ./cmd/api
	@echo "Done!"

build_image_proccessing_service:
	@echo "Building image_proccessing binary..."
	cd ../image_proccessing_service && docker build -t image-proccessing-service -f image-proccessing-service.dockerfile .
	@echo "Done!"