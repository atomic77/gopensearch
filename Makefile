BINARY_NAME=gopensearch

build:
	GOARCH=amd64 GOOS=linux go build -o ${BINARY_NAME}-linux-amd64 --tags 'fts5' ./cmd/gopensearch

static:
	GOARCH=amd64 GOOS=linux go build -o ${BINARY_NAME}-linux-amd64 -ldflags '-linkmode external -extldflags "-static"' --tags 'fts5' ./cmd/gopensearch

arm:
	GOARCH=arm GOOS=linux go build -o ${BINARY_NAME}-linux-armv7 --tags 'fts5' ./cmd/gopensearch
	GOARCH=arm64 GOOS=linux go build -o ${BINARY_NAME}-linux-arm64 --tags 'fts5' ./cmd/gopensearch

test:
	go test ./... --tags 'fts5'

clean:
	rm ${BINARY_NAME}-linux
