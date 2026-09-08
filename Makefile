.PHONY: help test fmt vet install-deps docker-setup docker-down setup down clean init docker-init

# Labels for help output:
#   [host]      - Run from the host machine (outside the dev container)
#   [container] - Run from inside the dev container
#   [any]       - Can be run from either environment

## help: Show this help message
help:
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "  [host]      targets must be run on the host machine (they manage Docker Compose)"
	@echo "  [container] targets must be run inside the dev container"
	@echo "  [any]       targets can be run from either environment"
	@echo ""
	@awk '/^## / { \
		line = substr($$0, 4); \
		if (line ~ /^──/) { printf "\n%s\n", line; } \
		else { n = index(line, ": "); printf "  %-48s %s\n", substr(line,1,n-1), substr(line,n+2); } \
	}' $(MAKEFILE_LIST)
	@echo ""

OUTPUT=bin
APP=api

# Guard: verifies the target is being executed inside the dev container.
# Set HOST_ALT as a target-specific variable to surface the host alternative.
define REQUIRE_CONTAINER
@if [ "$(INSIDE_DEV_CONTAINER)" != "true" ]; then \
	printf '\nError: "%s" must be run inside the dev container.\n' "$@"; \
	if [ -n "$(HOST_ALT)" ]; then \
		printf '       From the host, run:  make %s\n' "$(HOST_ALT)"; \
	fi; \
	printf 'Tip:   Enter the container with: docker compose exec app bash\n\n'; \
	exit 1; \
fi
endef

## ── Dev environment ──────────────────────────────────────────────────────────

## setup [host]: Start the dev environment (Docker Compose)
setup: docker-setup
	@echo "Dev environment is up and running."

## teardown [host]: Remove the dev environment (Docker Compose)
teardown: docker-destroy
	@echo "Dev environment has been removed."

## down [host]: Stop the dev environment (Docker Compose)
down: docker-down
	@echo "Dev environment has been stopped."

## docker-setup [host]: Start the app container via Docker Compose
docker-setup:
	docker compose up -d app

## docker-down [host]: Stop all containers via Docker Compose
docker-down:
	docker compose down

## docker-destroy [host]: Remove all containers, images, and volumes via Docker Compose
docker-destroy:
	docker compose down --rmi all --volumes --remove-orphans

## docker-shell [host]: Open a shell inside the app container
docker-shell: docker-setup
	docker compose exec app zsh

## ── Build & test (host wrappers) ─────────────────────────────────────────────

## docker-test [host]: Run tests inside the dev container
docker-test: docker-setup
	docker compose exec app make test

## docker-build [host]: Build binaries inside the dev container
docker-build: docker-setup
	docker compose exec app make build

docker-fmt: docker-setup
	docker compose exec app make fmt

docker-vet: docker-setup
	docker compose exec app make vet

docker-install-deps: docker-setup
	docker compose exec app make install-deps

docker-swagger-gen: docker-setup
	docker compose exec app make swagger-gen

## docker-init [host]: Rename the module path inside the dev container (usage: make docker-init MODULE=github.com/you/project)
docker-init: docker-setup
	docker compose exec app make init MODULE=$(MODULE)

## ── Build & test (inside container) ─────────────────────────────────────────

## test [container]: Run all tests
test: HOST_ALT = docker-test
test:
	$(REQUIRE_CONTAINER)
	go test ./...

## build [container]: Build the binary for $(APP)
build: HOST_ALT = docker-build
build:
	$(REQUIRE_CONTAINER)
	CGO_ENABLED=0 go build -o $(OUTPUT)/$(APP) ./cmd/$(APP)

## fmt [container]: Format all Go source files
fmt: HOST_ALT = docker-fmt
fmt:
	$(REQUIRE_CONTAINER)
	go fmt ./...

## vet [container]: Run go vet on all packages
vet: HOST_ALT = docker-vet
vet:
	$(REQUIRE_CONTAINER)
	go vet ./...

## install-deps [container]: Download Go module dependencies
install-deps: HOST_ALT = docker-install-deps
install-deps:
	$(REQUIRE_CONTAINER)
	go mod download

## init [container]: Rename the module path (usage: make init MODULE=github.com/you/project)
init: HOST_ALT = docker-init MODULE=github.com/you/project
init:
	$(REQUIRE_CONTAINER)
	@if [ -z "$(MODULE)" ]; then \
		echo "Error: MODULE is required, e.g. make init MODULE=github.com/you/project"; \
		exit 1; \
	fi
	@OLD=$$(go list -m); \
	grep -rl "$$OLD" --include='*.go' . | xargs -r sed -i.bak "s#$$OLD#$(MODULE)#g"; \
	find . -name '*.bak' -delete
	go mod edit -module $(MODULE)
	go mod tidy
	@echo "Module renamed to $(MODULE)"

## ── Run (inside container) ───────────────────────────────────────────────────

## run-api [container]: Start the API service with live-reload
run-api:
	$(REQUIRE_CONTAINER)
	air --build.cmd "go build -o ./tmp/main ./cmd/api"

## clean [any]: Remove build artifacts
clean:
	rm -rf bin/
	rm -rf tmp/
	rm -rf docs/
