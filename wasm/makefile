# Variables
DIST := dist
WASM := $(DIST)/bstates.wasm
LOADER := $(DIST)/wasm_exec.js

TINYGOROOT := $(shell tinygo env TINYGOROOT)
TINYGO_WASM_EXEC := $(TINYGOROOT)/targets/wasm_exec.js
TINYGO_BUILD_CMD := GOOS=js GOARCH=wasm tinygo build -o $(WASM) -tags noasm --no-debug

GOROOT := $(shell go env GOROOT)
GO_WASM_EXEC := $(GOROOT)/misc/wasm/wasm_exec.js
GO_BUILD_CMD := GOOS=js GOARCH=wasm go build -o $(WASM) -tags noasm

# Default target
default: dist gobuild go_wasm_exec

tinygo: dist tinygobuild tinygo_wasm_exec

# Create the $(DIST) directory
dist:
	@echo "Creating $(DIST) directory..."
	mkdir -p $(DIST)
	@echo "Directory created: $(DIST)"

# Compile the wasm file using go
gobuild:
	@echo "Compiling the WASM file..."
	$(GO_BUILD_CMD)
	@echo "File $(WASM) generated."
	@echo "Generating JS module..."
	node generate-wasm-module.js

# Copy go wasm_exec.js
go_wasm_exec:
	@echo "Copying go wasm_exec.js..."
	cp -f $(GO_WASM_EXEC) $(LOADER)

# Compile the wasm file using tinygo
tinygobuild:
	@echo "Compiling the WASM file..."
	$(TINYGO_BUILD_CMD)
	@echo "File $(WASM) generated."
	@echo "Generating JS module..."
	node generate-wasm-module.js

# Copy tinygo wasm_exec.js
tinygo_wasm_exec:
	@echo "Copying tinygo wasm_exec.js..."
	cp -f $(TINYGO_WASM_EXEC) $(LOADER)

# Clean generated files
clean:
	@echo "Cleaning up generated files..."
	rm -f $(DIST)/*
	@echo "Files cleaned."
