export GO111MODULE=on
export PATH:=/usr/local/go/bin:$(PATH)
exec_path := /usr/local/bin/
exec_name := gophertelekomcloud


default: test
test: test-unit
acceptance: test-acc

fmt:
	@echo Running go fmt
	@go fmt

lint:
	@echo Running go lint
	@golangci-lint run --timeout=300s

vet:
	@echo "go vet ."
	@go vet ./...

test-unit:
	@go test ./openstack/... -parallel 4 -v

test-acc:
	@echo "Starting acceptance tests..."
	@go test ./acceptance/... -race -covermode=atomic -coverprofile=coverage.txt -timeout 20m -v
