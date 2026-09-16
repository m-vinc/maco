BINARY   := maco
PKG      := ./cmd/maco
VERSION  ?= dev
LDFLAGS  := -X main.VERSION=$(VERSION)
GOFLAGS  ?=

PROD ?= 0
ifeq ($(PROD),1)
GOTAGS := prod
UI_DEP := web-build
endif
TAGFLAGS := $(if $(GOTAGS),-tags $(GOTAGS),)

.DEFAULT_GOAL := build

FIRMWARE_ASSET := pkg/firmware/assets/maco-aarch64-code.fd.bz2
FIRMWARE_DEPS  := tools/firmware/Dockerfile tools/firmware/logo.py tools/firmware/strip-network.sh web/public/logo.svg

.PHONY: firmware-build
firmware-build: $(FIRMWARE_ASSET) ## build the embedded Maco-branded ARM64 UEFI firmware if stale (Docker)

$(FIRMWARE_ASSET): $(FIRMWARE_DEPS)
	docker build --file tools/firmware/Dockerfile --output type=local,dest=pkg/firmware/assets .
	touch $(FIRMWARE_ASSET)

.PHONY: firmware-rebuild
firmware-rebuild: ## force a rebuild of the Maco-branded UEFI firmware (Docker)
	docker build --no-cache --file tools/firmware/Dockerfile --output type=local,dest=pkg/firmware/assets .

export GOOS  ?= darwin
export GOARCH ?= arm64

.PHONY: build
build: maco-net-helper $(UI_DEP) ## build the maco binary (PROD=1 builds+embeds the web UI)
	go build $(GOFLAGS) $(TAGFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY) $(PKG)

.PHONY: build-ui
build-ui: ## build the maco binary with the embedded web UI (alias for build PROD=1)
	$(MAKE) build PROD=1

.PHONY: web-build
web-build: maco-net-helper ## build the React/Vite frontend into pkg/api/dist
	cd web && npm install && npm run build

.PHONY: web-dev
web-dev: ## run the Vite dev server (HMR) proxying /api to a local maco serve
	cd web && npm install && npm run dev

.PHONY: sqlc
sqlc: ## regenerate type-safe DB code from pkg/db/query + migrations
	sqlc generate

.PHONY: test
test: test-network-helper ## vet + unit tests
	go vet ./...
	go test ./...

.PHONY: lint
lint: ## run golangci-lint
	golangci-lint run ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: run
run: build ## build then run (pass ARGS="vm list")
	./$(BINARY) $(ARGS)

PKG_ID       := com.maco.pkg
PKG_STAGING  := build/pkgroot
PKG_SCRIPTS  := build/scripts

.PHONY: pkg
pkg: build-ui ## build an unsigned macOS installer .pkg (VERSION=x.y.z); requires Redis via Homebrew
	rm -rf build maco-$(VERSION).pkg
	mkdir -p $(PKG_STAGING)/usr/local/bin $(PKG_SCRIPTS)
	cp $(BINARY) $(PKG_STAGING)/usr/local/bin/maco
	cp tools/pkg/preinstall $(PKG_SCRIPTS)/preinstall
	cp tools/pkg/postinstall $(PKG_SCRIPTS)/postinstall
	chmod +x $(PKG_SCRIPTS)/preinstall $(PKG_SCRIPTS)/postinstall
	xattr -cr $(PKG_STAGING)
	pkgbuild --root $(PKG_STAGING) --scripts $(PKG_SCRIPTS) --identifier $(PKG_ID) --version $(VERSION) --install-location / build/maco-component.pkg
	productbuild --package build/maco-component.pkg maco-$(VERSION).pkg
	@echo "built maco-$(VERSION).pkg"

.PHONY: clean
clean:
	rm -rf build maco-*.pkg
	rm -f $(BINARY) maco-net-helper

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

HELPER_ASSET := pkg/nethelper/assets/maco-net-helper

.PHONY: maco-net-helper
maco-net-helper:
	mkdir -p $(dir $(HELPER_ASSET))
	$(CC) -O2 -Wall -Wextra -Werror pkg/l2/native/helper.c -o $(HELPER_ASSET)
	cp $(HELPER_ASSET) maco-net-helper

.PHONY: test-network-helper
test-network-helper:
	$(CC) -g -Wall -Wextra -Werror -Wno-unused-function -fsanitize=address,undefined tools/l2-validation/batching-test.c -o /private/tmp/maco-batching-test
	/private/tmp/maco-batching-test

SWAG_VERSION := v2.0.0-rc5

.PHONY: docs api-tools api-generate api-check
docs: api-generate ## generate the API contract from Go code

api-tools: ## install the pinned Go OpenAPI generator
	go install github.com/swaggo/swag/v2/cmd/swag@$(SWAG_VERSION)

api-generate: ## regenerate OpenAPI 3.1 and frontend client types from Go handlers
	swag init --v3.1 -g doc.go -d ./pkg/api --parseDependency --parseInternal --requiredByDefault --outputTypes json -o pkg/api/docs
	python3 tools/openapi/normalize.py pkg/api/docs/swagger.json
	cd web && npm run api:generate

api-check: ## verify the committed spec/client match Go handlers and mounted routes
	sh tools/openapi/check.sh
