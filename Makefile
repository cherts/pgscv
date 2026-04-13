DOCKER_ACCOUNT = cherts
DOCKER_BUILD_PLATFORM ?= linux/amd64
DOCKER_PLATFORMS ?= linux/amd64,linux/arm64
DOCKER_BUILDX_BUILDER ?= pgscv-builder
APPNAME = pgscv

TAG_COMMIT := $(shell git rev-list --abbrev-commit --tags --max-count=1)
TAG := $(shell git describe --abbrev=0 --tags ${TAG_COMMIT} 2>/dev/null || true)
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell git log -1 --format=%cd --date=format:"%Y%m%d")
ifeq ($(TAG),)
	VERSION := 0.15
else
	#VERSION := $(TAG:v%=%)
	VERSION := $(TAG)
endif
ifneq ($(COMMIT), $(TAG_COMMIT))
    VERSION := $(VERSION)-next-$(DATE)
endif
ifeq ($(VERSION),)
    VERSION := $(COMMIT)-$(DATA)
endif
ifneq ($(shell git status --porcelain),)
    VERSION := $(VERSION)-dirty
endif
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

VERSION_BETA=$(BRANCH)-$(COMMIT)-$(DATE)-beta
SANITIZED_BETA_TAG := $(subst release/0.15,0.15,$(VERSION_BETA))
SANITIZED_BETA_TAG := $(subst /,-,$(SANITIZED_BETA_TAG))

LDFLAGS = -a -installsuffix cgo -ldflags "-X main.appName=${APPNAME} -X main.gitTag=${VERSION} -X main.gitCommit=${COMMIT} -X main.gitBranch=${BRANCH}"
LDFLAGS_BETA = -a -installsuffix cgo -ldflags "-X main.appName=${APPNAME} -X main.gitTag=${SANITIZED_BETA_TAG} -X main.gitCommit=${COMMIT} -X main.gitBranch=${BRANCH}"

MODERNIZE_CMD = go run golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest

.PHONY: help \
		clean lint test race \
		build docker-lint docker-buildx-setup docker-build docker-push go-update \
		modernize modernize-fix modernize-check

.DEFAULT_GOAL := help

help: ## Display this help screen
	@echo "Makefile available targets:"
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  * \033[36m%-25s\033[0m %s\n", $$1, $$2}'

clean: ## Clean
	rm -f ./bin/${APPNAME} ./bin/${APPNAME}.tar.gz ./bin/${APPNAME}.version ./bin/${APPNAME}.sha256
	rm -rf ./bin

go-update: ## Update go mod
	go mod tidy -compat=1.25
	go get -u ./cmd
	go mod download
	go get -u ./cmd
	go mod download

dep: ## Get the dependencies
	go mod download

lint: ## Lint the source files
	go env -w GOFLAGS="-buildvcs=false"
	go vet ./...
	REVIVE_FORCE_COLOR=1 revive -formatter friendly ./...
	gosec -quiet ./...

test: dep lint ## Run tests
	go test -race -timeout 300s -coverprofile=.test_coverage.txt ./... && \
    	go tool cover -func=.test_coverage.txt | tail -n1 | awk '{print "Total test coverage: " $$3}'
	@rm .test_coverage.txt

race: dep ## Run data race detector
	go test -race -short -timeout 300s -p 1 ./...

build: dep ## Build
	mkdir -p ./bin
	CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build ${LDFLAGS} -o bin/${APPNAME} ./cmd

build-beta: dep ## Build beta
	mkdir -p ./bin
	CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build ${LDFLAGS_BETA} -o bin/${APPNAME} ./cmd

docker-lint: ## Lint Dockerfile
	@echo "Lint container Dockerfile"
	docker run --rm -i -v $(PWD)/Dockerfile:/Dockerfile \
	hadolint/hadolint hadolint --ignore DL3002 --ignore DL3008 --ignore DL3059 /Dockerfile

docker-buildx-setup: ## Setup docker buildx builder
	docker buildx inspect $(DOCKER_BUILDX_BUILDER) > /dev/null || docker buildx create --name $(DOCKER_BUILDX_BUILDER)
	docker buildx use $(DOCKER_BUILDX_BUILDER)
	docker buildx inspect --bootstrap > /dev/null

docker-build: docker-buildx-setup ## Build docker image
	docker buildx build --no-cache --platform $(DOCKER_BUILD_PLATFORM) --tag ${DOCKER_ACCOUNT}/${APPNAME}:${TAG} --file ./Dockerfile --load .

docker-push: ## Push docker image
	docker buildx build --no-cache --platform $(DOCKER_PLATFORMS) --tag $(DOCKER_ACCOUNT)/${APPNAME}:${TAG} --tag $(DOCKER_ACCOUNT)/${APPNAME}:latest --file ./Dockerfile --push .

docker-build-beta: docker-buildx-setup ## Build docker image (beta)
	docker buildx build --no-cache --platform $(DOCKER_BUILD_PLATFORM) --tag ${DOCKER_ACCOUNT}/${APPNAME}:v${SANITIZED_BETA_TAG} --file ./Dockerfile.beta --load .

docker-push-beta: ## Push docker image (beta)
	docker buildx build --no-cache --platform $(DOCKER_PLATFORMS) --tag $(DOCKER_ACCOUNT)/${APPNAME}:v${SANITIZED_BETA_TAG} --file ./Dockerfile.beta --push .

modernize: modernize-fix ## Run gopls modernize check and fix

modernize-fix: ## Run gopls modernize fix
	@echo "Running gopls modernize with -fix..."
	go env -w GOFLAGS="-buildvcs=false"
	$(MODERNIZE_CMD) -test -fix ./...

modernize-check: ## Run gopls modernize only check
	@echo "Checking if code needs modernization..."
	go env -w GOFLAGS="-buildvcs=false"
	$(MODERNIZE_CMD) -test ./...