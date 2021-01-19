include .bingo/Variables.mk

MODULES ?= $(shell find $(PWD) -name "go.mod" | grep -v ".bingo" | xargs dirname)

GO111MODULE       ?= on
export GO111MODULE

GOBIN ?= $(firstword $(subst :, ,${GOPATH}))/bin
TMP_DIR := /tmp/rndr

# Tools.
GIT ?= $(shell which git)

# Support gsed on OSX (installed via brew), falling back to sed. On Linux
# systems gsed won't be installed, so will use sed as expected.
SED ?= $(shell which gsed 2>/dev/null || which sed)

define require_clean_work_tree
	@git update-index -q --ignore-submodules --refresh

    @if ! git diff-files --quiet --ignore-submodules --; then \
        echo >&2 "$1: you have unstaged changes."; \
        git diff-files --name-status -r --ignore-submodules -- >&2; \
        echo >&2 "Please commit or stash them."; \
        exit 1; \
    fi

    @if ! git diff-index --cached --quiet HEAD --ignore-submodules --; then \
        echo >&2 "$1: your index contains uncommitted changes."; \
        git diff-index --cached --name-status -r --ignore-submodules HEAD -- >&2; \
        echo >&2 "Please commit or stash them."; \
        exit 1; \
    fi

endef

help: ## Displays help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build rndr.
	@echo ">> building rndr"
	@go build -o $(GOBIN)/rndr ./cmd/rndr/...

.PHONY: deps
deps: ## Cleans up deps for all modules
	@echo ">> running deps tidy for all modules: $(MODULES)"
	for dir in $(MODULES) ; do \
		cd $${dir} && go mod tidy; \
	done

.PHONY: format
format: ## Formats Go code.
format: $(GOIMPORTS)
	@echo ">> formatting  all modules Go code: $(MODULES)"
	@$(GOIMPORTS) -w $(MODULES)

.PHONY: test
test: ## Runs all Go unit tests.
test: test-examples
	@echo ">> running tests for all modules: $(MODULES)"
	for dir in $(MODULES) ; do \
		cd $${dir} && go test -v -race ./...; \
	done

.PHONY: test-examples
test-examples: ## Test examples.
test-examples: build
	@$(MAKE) -C "examples/hellosvc" kubernetes
	@$(MAKE) -C "examples/hellosvc" kubernetes-special

# Tooling for jsonnet examples.
JSONNET_SRC = $(shell find . -type f -not -path './*vendor/*' \( -name '*.libsonnet' -o -name '*.jsonnet' \))
EXAMPLE_JSONNET_HELLOSVC_DIR="examples/hellosvc/hellosvc-tmpl-jsonnet"

.PHONY: jsonnet-example
jsonnet-example: # Generate example service.
jsonnet-example: $(JB) $(JSONNETFMT) $(JSONNET_LINT) $(GOJSONTOYAML) $(JSONNET)
	@echo "format all"
	@$(JSONNETFMT) -n 2 --max-blank-lines 2 --string-style s --comment-style s -i $(JSONNET_SRC)
	@echo "lint all"
	@echo ${JSONNET_SRC} | xargs -n 1 -- $(JSONNET_LINT) -J vendor
	@echo "install deps for $(EXAMPLE_JSONNET_HELLOSVC_DIR)"
	@cd $(EXAMPLE_JSONNET_HELLOSVC_DIR) && $(JB) install
	@echo "generate Kubernetes service manually using jsonnet to check if we can for $(EXAMPLE_JSONNET_HELLOSVC_DIR)"
	@mkdir -p $(TMP_DIR)/$(EXAMPLE_JSONNET_HELLOSVC_DIR)
	@echo "(import '$(shell pwd)/$(EXAMPLE_JSONNET_HELLOSVC_DIR)/hellosvc.libsonnet')({})" > $(TMP_DIR)/$(EXAMPLE_JSONNET_HELLOSVC_DIR)/main.jsonnet
	$(JSONNET) -J vendor -m $(TMP_DIR)/$(EXAMPLE_JSONNET_HELLOSVC_DIR) $(TMP_DIR)/$(EXAMPLE_JSONNET_HELLOSVC_DIR)/main.jsonnet | xargs -I{} sh -c 'cat {} | $(GOJSONTOYAML)' -- {}

.PHONY: check-git
check-git:
ifneq ($(GIT),)
	@test -x $(GIT) || (echo >&2 "No git executable binary found at $(GIT)."; exit 1)
else
	@echo >&2 "No git binary found."; exit 1
endif

# PROTIP:
# Add
#      --cpu-profile-path string   Path to CPU profile output file
#      --mem-profile-path string   Path to memory profile output file
# to debug big allocations during linting.
lint: ## Runs various static analysis against our code.
lint: $(FAILLINT) $(GOLANGCI_LINT) $(MISSPELL) build format check-git deps
	$(call require_clean_work_tree,"detected not clean master before running lint - run make lint and commit changes.")
	@echo ">> verifying imported "
	for dir in $(MODULES) ; do \
		cd $${dir} && $(FAILLINT) -paths "fmt.{Print,PrintfPrintln,Sprint}" -ignore-tests ./...; \
	done
	@echo ">> examining all of the Go files"
	for dir in $(MODULES) ; do \
		cd $${dir} && go vet -stdmethods=false ./...; \
	done
	@echo ">> linting all of the Go files GOGC=${GOGC}"
	for dir in $(MODULES) ; do \
		cd $${dir} && $(GOLANGCI_LINT) run; \
	done
	@echo ">> detecting misspells"
	@find . -type f | grep -v vendor/ | grep -vE '\./\..*' | xargs $(MISSPELL) -error
	$(call require_clean_work_tree,"found changes, run make lint and commit changes.")
