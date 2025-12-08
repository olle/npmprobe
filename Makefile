ifeq ($(shell command -v podman 2> /dev/null),)
    CNTR := docker
else
    CNTR := podman
endif


.PHONY: all
all: build prepare-list
	./bin/npmprobe compromised.txt

##
## Normalize list of compromised packages by sorting and removing
## duplicate entries. Use this target before updating the repository.
##
.PHONY: prepare-list
prepare-list:
	cat compromised.txt | sort | uniq > temp.txt
	cp temp.txt compromised.txt
	rm temp.txt

##
## Cross-compile the npmprobe Go binary for multiple platforms.
## Outputs to bin/ directory with platform-specific names.
##
.PHONY: build
build: fmt verify
	@mkdir -p bin
	@echo "Building npmprobe for multiple platforms..."
	GOOS=darwin GOARCH=arm64 go build -o bin/npmprobe ./cmd/npmprobe
	GOOS=linux GOARCH=amd64 go build -o bin/npmprobe-linux-x86_64 ./cmd/npmprobe
	@echo "Build complete. Binaries in bin/"
	@ls -lh bin/npmprobe-*

##
## Format all Go source files.
##
.PHONY: format fmt
format fmt:
	go fmt ./...

##
## Run all tests to verify the codebase.
##
.PHONY: verify v
verify v:
	go test ./...

##
## Build and run the Docker test image (Linux)
##
.PHONY: docker-test
docker-test:
	@echo "Building Docker image 'npmprobe:test'..."
	${CNTR} build -t npmprobe:test .
	@echo "Running Docker image 'npmprobe:test'..."
	${CNTR} run --rm npmprobe:test
