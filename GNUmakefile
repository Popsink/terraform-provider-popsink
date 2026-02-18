PKG_NAME=popsink
TEST=./internal/...
COVERAGEARGS=-race -coverprofile=coverage.txt -covermode=atomic

# VARIABLE REFERENCE:
#
#   T=<pattern>       - Test name pattern (e.g., TestAccConnectorResource)
#   COV=true          - Enable coverage
#
# Examples:
#   make test                                # Run all unit tests
#   make test T=TestClient                   # Run only matching unit tests
#   make test COV=true                       # Run unit tests with coverage
#   make testacc T=TestAccConnector COV=true # Run specific acceptance tests with coverage

ifneq ($(origin T), undefined)
	RUNARGS = -run='$(T)'
endif

ifneq ($(origin COV), undefined)
	RUNARGS += $(COVERAGEARGS)
endif

default: build

build: lint
	CGO_ENABLED=0 go build -ldflags="-s -w" ./...

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/popsink/popsink/1.0.0/$$(go env GOOS)_$$(go env GOARCH)
	cp terraform-provider-popsink ~/.terraform.d/plugins/registry.terraform.io/popsink/popsink/1.0.0/$$(go env GOOS)_$$(go env GOARCH)/

fmt:
	@echo "==> Fixing source code formatting..."
	gofmt -s -w .

lint:
	@echo "==> Checking source code against linters..."
	golangci-lint run ./...

test:
	CGO_ENABLED=0 go test $(TEST) \
		-timeout=30s \
		-parallel=4 \
		-v \
		-skip '^TestAcc' \
		$(RUNARGS) \
		-count 1

testacc:
	TF_ACC=1 CGO_ENABLED=0 go test $(TEST) \
		-v \
		-run '^TestAcc' \
		$(RUNARGS) \
		-timeout 120m \
		-count=1

docs:
	go generate ./...

clean:
	rm -f terraform-provider-popsink
	rm -f coverage.txt
	go clean -testcache

.PHONY: build install fmt lint test testacc docs clean
