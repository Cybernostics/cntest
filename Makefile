PROJECTNAME=$(shell basename "$(PWD)")

GO   := go

# Go related variables.
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)
GOPKGS=$(shell go list -f {{.Dir}} ./... )

BIN = $(GOBASE)/bin
$(BIN):
	@mkdir -p $@

# COLORS
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)

#-------------------------
# Target: install package
#-------------------------
$(BIN)/%: | $(BIN)
	@tmp=$$(mktemp -d); \
		env GO111MODULE=off GOPATH=$$tmp GOBIN=$(BIN) go get $(PACKAGE) \
		|| ret=$$?; \
	   rm -rf $$tmp ; exit $$ret


## Apply code format, import reorganization and code simplification on source code
format:
	@echo "==> formatting code"
	@$(GO) fmt $(pkgs)
	@echo "==> clean imports"
	@goimports -w $(pkgDirs)
	@echo "==> simplify code"
	@gofmt -s -w $(pkgDirs)


get.tools:
	# License checker
	go get -u github.com/frapposelli/wwhrd
 	# linter
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

#-------------------------
# Target: lint
#-------------------------
$(BIN)/golint: PACKAGE=github.com/golangci/golangci-lint/cmd/golangci-lint

GOLINT = $(BIN)/golangci-lint
## Lint files
lint: | $(GOLINT)
	$(GOLINT) -vvv -set_exit_status ./...


#-------------------------
# Target: depend
#-------------------------
.PHONY: depend vendor.check depend.status depend.update depend.cleanlock depend.update.full

## Use go modules
depend: depend.tidy depend.verify

depend.tidy:
	@echo "==> Running dependency cleanup"
	$(GO) mod tidy -v

depend.verify:
	@echo "==> Verifying dependencies"
	$(GO) mod verify

depend.update:
	@echo "==> Update go modules"
	$(GO) get -u -v

depend.update.full: depend.cleanlock depend.update

#-------------------------
# Target: clean
#-------------------------

## Clean build files
clean: clean.go
	rm -rf $(DIRS_TO_CLEAN)
	rm -f $(FILES_TO_CLEAN)

clean.go: ; $(info cleaning...)
	$(eval GO_CLEAN_FLAGS := -i -r)
	$(GO) clean $(GO_CLEAN_FLAGS)	

TARGET_MAX_CHAR_NUM=20
#-------------------------
# Target: help
#-------------------------

## Show help
help:
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  ${YELLOW}%-$(TARGET_MAX_CHAR_NUM)s${RESET} ${GREEN}%s${RESET}\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)	

	#-------------------------
# Code generation
#-------------------------
.PHONY: generate

## Generate go code
generate:
	@echo "==> generating go code"
	$(GO) generate $(pkgs)

#-------------------------
# Build artefacts
#-------------------------
.PHONY: build build.http-server-go

## Build all binaries
build:
	$(GO) build $(pkgs)

test: build
	$(GO) test $(pkgs)

