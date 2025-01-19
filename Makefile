.PHONY: all
all: lint test

.PHONY: tidy
tidy:
	go mod tidy
	go mod -C internal/tools tidy

.PHONY: tools
tools:
	go install -C internal/tools \
		github.com/fdaines/spm-go \
		github.com/golangci/golangci-lint/cmd/golangci-lint \
		github.com/google/yamlfmt/cmd/yamlfmt \
		github.com/jackc/tern/v2 \
		github.com/maxbrunsfeld/counterfeiter/v6 \
		github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		github.com/sqlc-dev/sqlc/cmd/sqlc \
		goa.design/model/cmd/mdl \
		goa.design/model/cmd/stz \
		golang.org/x/vuln/cmd/govulncheck

# Formatting

.PHONY: fmt
fmt:
	go fmt ./...
	yamlfmt .

.PHONY: dirty
dirty:
	@status=$$(git status --untracked-files=no --porcelain); \
	if [ ! -z "$${status}" ]; \
	then \
		echo "ERROR: Working directory contains modified files"; \
		git status --untracked-files=no --porcelain; \
		exit 1; \
	fi

# Generate

.PHONY: generate
generate:
	go generate ./...

# Lint

.PHONY: lint
lint: tidy tools fmt security
	golangci-lint run ./...
	go vet ./...

# Security

.PHONY: security
security: 
	govulncheck ./...

# Test

.PHONY: test
test:
	go test -shuffle=on -race -coverprofile=coverage.txt -covermode=atomic $$(go list ./... | grep -v /cmd/)


# Build
.PHONY: docker
docker:
	docker compose -f compose.yml build
