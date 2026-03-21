.PHONY: help build run test clean docker docker-run deploy

help:
	@echo "VM Manager - Mission Control Phase 2"
	@echo ""
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  run          - Run locally"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker       - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  deploy       - Deploy to Fly.io"
	@echo "  logs         - Tail Fly.io logs"

build:
	@echo "🔨 Building vm-manager..."
	go build -o vm-manager main.go
	@echo "✅ Build complete: ./vm-manager"

run:
	@echo "🚀 Running vm-manager..."
	go run main.go

test:
	@echo "🧪 Running tests..."
	go test -v -race ./...

clean:
	@echo "🧹 Cleaning..."
	rm -f vm-manager
	go clean

docker:
	@echo "🐳 Building Docker image..."
	docker build -t mission-control-vm-manager:latest .
	@echo "✅ Image ready: mission-control-vm-manager:latest"

docker-run:
	@echo "🐳 Running Docker container..."
	docker run -p 8080:8080 \
		-e FLY_API_TOKEN="${FLY_API_TOKEN}" \
		-e FLY_APP="mission-control-vms" \
		--name vm-manager \
		mission-control-vm-manager:latest

docker-stop:
	docker stop vm-manager && docker rm vm-manager

deploy:
	@echo "🚀 Deploying to Fly.io..."
	fly deploy

logs:
	@echo "📋 Tailing Fly.io logs..."
	fly logs

fmt:
	@echo "✨ Formatting code..."
	go fmt ./...

lint:
	@echo "🔍 Linting code..."
	golangci-lint run --timeout 5m

vet:
	@echo "🔍 Running go vet..."
	go vet ./...

deps:
	@echo "📦 Downloading dependencies..."
	go mod download
	go mod tidy

.PHONY: help build run test clean docker docker-run docker-stop deploy logs fmt lint vet deps
