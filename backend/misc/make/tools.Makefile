# This makefile should be used to hold functions/variables

ifeq ($(ARCH),x86_64)
	ARCH := amd64
else ifeq ($(ARCH),aarch64)
	ARCH := arm64
endif

define github_url
    https://github.com/$(GITHUB)/releases/download/v$(VERSION)/$(ARCHIVE)
endef

# creates a directory bin.
bin:
	@ mkdir -p $@

# ~~~ Tools ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

# ~~ [migrate] ~~~ https://github.com/golang-migrate/migrate ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
# ~~ Link release ARM: https://github.com/golang-migrate/migrate/releases/download/v4.18.2/migrate.darwin-arm64.tar.gz

MIGRATE := $(shell command -v migrate || echo "bin/migrate")
migrate: bin/migrate ## Install migrate (database migration)

bin/migrate: VERSION := 4.18.3
bin/migrate: GITHUB  := golang-migrate/migrate
bin/migrate: ARCHIVE := migrate.$(OSTYPE)-$(ARCH).tar.gz
bin/migrate: bin
	@ printf "Install migrate...\n"
	@ printf "Download URL: $(call github_url)\n"
	@ curl -Ls $(shell echo $(call github_url) | tr A-Z a-z) | tar -zOxf - migrate > $@ && chmod +x $@
	@ printf "done.\n"

# ~~ [ air ] ~~~ https://github.com/cosmtrek/air ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
# ~~ Link release ARM: https://github.com/air-verse/air/releases/download/v1.61.7/air_1.61.7_darwin_amd64.tar.gz

AIR := $(shell command -v air || echo "bin/air")
air: bin/air ## Installs air (go file watcher)

bin/air: VERSION := 1.61.7
bin/air: GITHUB  := cosmtrek/air
bin/air: ARCHIVE := air_$(VERSION)_$(OSTYPE)_$(ARCH).tar.gz
bin/air: bin
	@ printf "Install air...\n"
	@ printf "Download URL: $(call github_url)\n"
	@ curl -Ls $(shell echo $(call github_url) | tr A-Z a-z) | tar -zOxf - air > $@ && chmod +x $@
	@ printf "done.\n"

# ~~ [ protoc ] ~~~ https://github.com/protocolbuffers/protobuf ~~~~~~~~~~~~~~~~~~~~~~~~~
PROTOC := $(shell command -v protoc || echo "bin/protoc")
PROTOC_ARCH := $(ARCH)
ifeq ($(ARCH),amd64)
	PROTOC_ARCH := x86_64
else ifeq ($(ARCH),arm64)
	PROTOC_ARCH := aarch_64
endif

PROTOC_OSTYPE := $(OSTYPE)
ifeq ($(OSTYPE),darwin)
	PROTOC_OSTYPE := osx
endif

# ~~ [ protoc-gen-validate ] ~~~ https://github.com/bufbuild/protovalidate/archive/refs/tags/v0.10.1.tar.gz ~~~~~~~~

protoc: bin/protoc ## Installs protobuf (grpc code generation)

bin/protoc: VERSION := 29.3
bin/protoc: GITHUB  := protocolbuffers/protobuf
bin/protoc: ARCHIVE := protoc-${VERSION}-$(PROTOC_OSTYPE)-$(PROTOC_ARCH).zip
bin/protoc: bin
	@ printf "Install protobuf...\n"
	@ mkdir -p ./bin/protovalidate
	@ echo "Downloading protovalidate..."
	@ curl -Ls "https://github.com/bufbuild/protovalidate/archive/refs/tags/v0.10.1.tar.gz" -o ./bin/protovalidate.tar.gz
	@ tar -xzf ./bin/protovalidate.tar.gz --strip-components=1 -C ./bin/protovalidate
	@ rm -f ./bin/protovalidate.tar.gz

	@ echo "Downloading protoc..."
	@ curl -Ls $(shell echo $(call github_url) | tr A-Z a-z) -o ./bin/protoc.zip
	@ cd ./bin && unzip -q protoc.zip && mv bin/protoc ./protoc && chmod +x protoc && rm -rf ./bin readme.txt protoc.zip
	@ printf "done.\n"

# ~~ [ go-enum ] ~~~ https://github.com/abice/go-enum ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
# ~~ Link release ARM: https://github.com/abice/go-enum/releases/download/v0.9.1/go-enum_Darwin_arm64

GO_ENUM := $(shell command -v go-enum || echo "bin/go-enum")
go-enum: bin/go-enum ## Installs go-enum (go file watcher)

bin/go-enum: VERSION := 0.9.1
bin/go-enum: GITHUB  := abice/go-enum
bin/go-enum: ARCHIVE := go-enum_$(shell uname -s)_$(shell uname -m)
bin/go-enum: bin
	@ printf "Install go-enum...\n"
	@ printf "Download URL: $(call github_url)\n"
	@ curl -L -o ./bin/go-enum $(call github_url) && chmod +x $@
	@ printf "done.\n"
