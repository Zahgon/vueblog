.PHONY: build test vet fmt run cover clean

build:
	go build ./...

test:
	go test ./... -count=1

vet:
	go vet ./...

fmt:
	gofmt -l .

cover:
	go test ./... -coverpkg=./internal/... -coverprofile=coverage.out -count=1
	go tool cover -func=coverage.out | tail -1

run:
	go run ./cmd/vueblog

clean:
	rm -f coverage.out vueblog
