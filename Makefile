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
build:
	@mkdir -p bin
	@echo "Building npmprobe for multiple platforms..."
	GOOS=darwin GOARCH=arm64 go build -o bin/npmprobe ./cmd/npmprobe
	GOOS=darwin GOARCH=arm64 go build -o bin/npmprobe-darwin-aarch64 ./cmd/npmprobe
	GOOS=linux GOARCH=amd64 go build -o bin/npmprobe-linux-x86_64 ./cmd/npmprobe
	GOOS=windows GOARCH=amd64 go build -o bin/npmprobe-windows-x86_64.exe ./cmd/npmprobe
	@echo "Build complete. Binaries in bin/"
	@ls -lh bin/npmprobe-*
