build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags=release -ldflags="-s -w" -o bin/apcli ./cmd

run:
	go run ./cmd
