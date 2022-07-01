GIT_VERSION ?= $(shell git describe --tags --always --dirty)
GIT_HASH ?= $(shell git rev-parse HEAD)

PKG=sigs.k8s.io/release-utils/version
LDFLAGS=-X $(PKG).gitVersion=$(GIT_VERSION)
KO_PREFIX ?= ghcr.io/rikatz

.PHONY: build
build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o ./kubepug .

.PHONY: test
test:
	go test ./...

.PHONY: ko
ko:
	LDFLAGS="$(LDFLAGS)" GIT_HASH=$(GIT_HASH) GIT_VERSION=$(GIT_VERSION) \
	KO_DOCKER_REPO=${KO_PREFIX}/kubepug ko publish --bare \
		--platform=all \
		github.com/rikatz/kubepug

.PHONY: release
release:
	LDFLAGS="$(LDFLAGS)" goreleaser release

# used when need to validate the goreleaser
.PHONY: snapshot
snapshot:
	LDFLAGS="$(LDFLAGS)" goreleaser release --skip-sign --skip-publish --snapshot --rm-dist

.PHONY: clean
clean:
	rm -rf kubepug
	rm -rf dist/
