# Variables
OUTPUT := bstates.wasm

TINYGOROOT := $(shell tinygo env TINYGOROOT)
TINYGO_WASM_EXEC := $(TINYGOROOT)/targets/wasm_exec.js
TINYGO_BUILD_CMD := GOOS=js GOARCH=wasm tinygo build -o $(OUTPUT) -tags noasm --no-debug

GOROOT := $(shell go env GOROOT)
GO_WASM_EXEC := $(GOROOT)/misc/wasm/wasm_exec.js
GO_BUILD_CMD := GOOS=js GOARCH=wasm go build -o $(OUTPUT) -tags noasm

# Default target
default: gobuild golink

tinygo: tinygobuild tinygolink

# Compile the wasm file using go
gobuild:
	@echo "Compiling the WASM file..."
	$(GO_BUILD_CMD)
	@echo "File $(OUTPUT) generated."

# Create a symbolic link to go wasm_exec.js
golink:
	@echo "Creating a symbolic link for wasm_exec.js..."
	ln -sf $(GO_WASM_EXEC) ./wasm_exec.js
	@echo "Symbolic link created: wasm_exec.js -> $(GO_WASM_EXEC)"

# Compile the wasm file using tinygo
tinygobuild:
	@echo "Compiling the WASM file..."
	$(TINYGO_BUILD_CMD)
	@echo "File $(OUTPUT) generated."

# Create a symbolic link to tinygo wasm_exec.js
tinygolink:
	@echo "Creating a symbolic link for wasm_exec.js..."
	ln -sf $(TINYGO_WASM_EXEC) ./wasm_exec.js
	@echo "Symbolic link created: wasm_exec.js -> $(TINYGO_WASM_EXEC)"

# Clean generated files
clean:
	@echo "Cleaning up generated files..."
	rm -f $(OUTPUT) wasm_exec.js
	@echo "Files cleaned."
