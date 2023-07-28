GIT_VERSION ?= $(shell git describe --tags --always --dirty)
GIT_HASH ?= $(shell git rev-parse HEAD)

PKG=sigs.k8s.io/release-utils/version
LDFLAGS=-X $(PKG).gitVersion=$(GIT_VERSION)
KO_PREFIX ?= ghcr.io/rikatz

.PHONY: build
build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o ./output/kubepug .

.PHONY: test
test:
	go test ./... -coverprofile coverage.out

.PHONY: ko
ko:
	LDFLAGS="$(LDFLAGS)" GIT_HASH=$(GIT_HASH) GIT_VERSION=$(GIT_VERSION) \
	KO_DOCKER_REPO=${KO_PREFIX}/kubepug ko publish --bare --tags latest --tags $(GIT_VERSION) \
		--platform=all --image-refs kubepugImagerefs \
		github.com/rikatz/kubepug

.PHONY: release
release:
	LDFLAGS="$(LDFLAGS)" goreleaser release --clean

# used when need to validate the goreleaser
.PHONY: snapshot
snapshot:
	LDFLAGS="$(LDFLAGS)" goreleaser release --skip-sign --skip-publish --snapshot --clean

.PHONY: clean
clean:
	rm -rf output/kubepug
	rm -rf dist/
