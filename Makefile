# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

test:
	go test ./... -coverprofile cover.out