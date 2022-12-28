BINARY_NAME=glim

build:
	go build -o main.go

run:
	./${BINARY_NAME}

build_and_run: build run

clean:
	go clean
	rm ${BINARY_NAME}

test:
	go test ./...

test_coverage:
	go test ./... -short -coverprofile=coverage.out `go list ./.. | grep -v vendor/`

dep:
	go mod download

vet:
	go vet