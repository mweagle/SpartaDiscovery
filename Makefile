.DEFAULT_GOAL=build
.PHONY: test delete explore provision describe build

validate:
	go tool vet *.go

format:
	go fmt .

build: format validate
	go build .
	@echo "Build complete"

test: build
	go test ./test/...

delete:
	go run main.go delete

explore:
	go run main.go --level info explore

provision:
	go run main.go --level info provision --s3Bucket $(S3_BUCKET)

describe: build
	go run main.go --level info describe --out ./graph.html
