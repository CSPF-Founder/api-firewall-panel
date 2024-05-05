build_frontend:
	cd frontend && npm install && npm run build

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags=release -ldflags="-s -w" -o bin/server ./cmd

# image:

run:
	go run ./cmd

test:
	go test -v ./...
	