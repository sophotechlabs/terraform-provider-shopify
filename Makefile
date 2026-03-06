GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BUMP ?= patch

PROVIDER_NAME := terraform-provider-shopify
PROVIDER_PATH := registry.terraform.io/sophotechlabs/shopify/0.0.1/$(GOOS)_$(GOARCH)

ifeq ($(GOOS),darwin)
	INSTALL_PATH := $(HOME)/Library/Application Support/io.terraform/plugins/$(PROVIDER_PATH)
else
	INSTALL_PATH := $(HOME)/.local/share/terraform/plugins/$(PROVIDER_PATH)
endif

.PHONY: default build test testacc lint vet fmt fmtcheck generate dev docs ci release

default: testacc

build:
	go build -v .

test:
	go test ./... -v

testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

lint:
	golangci-lint run ./...

vet:
	go vet ./...

fmt:
	gofmt -s -w .

fmtcheck:
	@sh -c 'files=$$(gofmt -s -l .); if [ -n "$$files" ]; then echo "Files need formatting:"; echo "$$files"; exit 1; fi'

generate:
	go generate ./...

dev: build
	@mkdir -p "$(INSTALL_PATH)"
	@cp $(PROVIDER_NAME) "$(INSTALL_PATH)/$(PROVIDER_NAME)"

docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate

ci: vet fmtcheck test

release:
	@LATEST=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	MAJOR=$$(echo $$LATEST | sed 's/^v//' | cut -d. -f1); \
	MINOR=$$(echo $$LATEST | sed 's/^v//' | cut -d. -f2); \
	PATCH=$$(echo $$LATEST | sed 's/^v//' | cut -d. -f3); \
	case "$(BUMP)" in \
		major) MAJOR=$$((MAJOR + 1)); MINOR=0; PATCH=0 ;; \
		minor) MINOR=$$((MINOR + 1)); PATCH=0 ;; \
		patch) PATCH=$$((PATCH + 1)) ;; \
	esac; \
	VERSION="v$$MAJOR.$$MINOR.$$PATCH"; \
	echo "$$LATEST → $$VERSION ($(BUMP) bump)"; \
	echo "Press Enter to tag and push, or Ctrl+C to abort"; \
	read _confirm; \
	git tag "$$VERSION" && git push origin "$$VERSION"
