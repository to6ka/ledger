
export VERSION := $(shell git describe --exact-match --tags)
export APP_SHA := $(shell git rev-parse --short=8 HEAD)

.PNONY: build
build:
	@echo "Building..."
	docker build \
		-t github.com/formancehq/ledger:$(VERSION) \
		--build-arg TARGETARCH=amd64 \
		--build-arg APP_SHA=$(APP_SHA) \
		--build-arg VERSION=$(VERSION) \
		.