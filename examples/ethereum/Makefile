VERSION ?= $(shell git describe --tags --always)
BUILD ?= $(shell git rev-parse --short HEAD)

# Go related variables.
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/bin

# Use linker flags to provide version/build settings
LDFLAGS_MOD="github.com/nakji-network/connector/config"
LDFLAGS=-ldflags "-X=$(LDFLAGS_MOD).Version=$(VERSION) -X=$(LDFLAGS_MOD).Build=$(BUILD) -X=$(LDFLAGS_MOD).BuildTime=$(shell date '+%s') -s -w" #remove debugger info for smaller binaries

PROTO_FILES = $(shell find . -type f -name '*.proto')
PBGO_FILES = $(patsubst %.proto,%.pb.go,$(PROTO_FILES))
GRPC_FILES = $(patsubst %.proto,%_grpc.pb.go,$(PROTO_FILES))
APPS = $(shell find . -path "./cmd/*/main.go" | sed -E "s/\\/main\.go//" | xargs -I {} basename {}) #all app names

.PHONY: all $(APPS) docker-all list help

## all: Build all apps
all: $(PBGO_FILES)
	go build -o bin ./...

%.pb.go : %.proto
	protoc --proto_path=. \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative $<

## start-<app>: Run app without compiling
start-%: $(PBGO_FILES)
	$(eval APATH := $(shell find . -path "./cmd/$*/main.go"))
	# commented out because I included the abi.go files in the repo
	# @test -f $(dir $(APATH))../init_connector.sh && \
		#echo "Running init_connector.sh" && \
		#$(dir $(APATH))../init_connector.sh || true
ifdef DYNAMIC
	go run -tags dynamic $(LDFLAGS) $(APATH)
else
	go run $(LDFLAGS) $(APATH)
endif

## <app>: Build app
$(APPS): $(PBGO_FILES)
	$(eval APATH := $(shell find . -path "./cmd/$@/main.go"))
	go build $(LDFLAGS) -o bin/$@ $(APATH)

## docker-all: Build docker image for all apps
docker: $(PGBO_FILES)
	docker build -f Dockerfile \
	--build-arg SSH_PRIVATE_KEY="$$SSH_PRIVATE_KEY" \
	#-t your.domain/ethereum-connector \
	#-t your.domain/ethereum-connector:$(BUILD) \
	.

## list: List all available apps
list:
	@echo $(APPS)

help: Makefile
	@echo
	@echo " Available commands:"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

clean:
	rm -f $(PBGO_FILES)
	rm -f $(GRPC_FILES)
	rm -rf $(GOBIN)

proto:
	protoc --proto_path=. \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative $(PROTO_FILES)

