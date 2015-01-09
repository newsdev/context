VERSION=0.0.1
OSARCH=darwin/amd64 linux/amd64 linux/386

OSARCH_BIN=$(OSARCH:%=builds/%/context)
OSARCH_BIN_GZ=$(OSARCH_BIN:%=%.gz)
OSARCH_BIN_GZ_CHECKSUM=$(OSARCH_BIN_GZ:%=%.sfv)

install:
	go install

release: $(OSARCH_BIN_GZ) $(OSARCH_BIN_GZ_CHECKSUM)
	aws s3 sync --delete builds s3://newsdev-pub/context/$(VERSION)

%.sfv: %
	shasum -a 512 $^ > $@

%.gz: %
	gzip -f -9 $^

$(OSARCH_BIN): builds
builds:
	gox -osarch '$(OSARCH)' -output 'builds/{{.OS}}/{{.Arch}}/context'

release-deps: gox deps

gox-toolchain: gox
	gox -build-toolchain -osarch $(OSARCH)

gox:
	go get -v github.com/mitchellh/gox

deps:
	go get -v -d ./...

clean:
	rm -rf builds

.PHONY: builds
.PRECIOUS: %.gz %.sfv
