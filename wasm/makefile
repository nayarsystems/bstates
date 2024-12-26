# Variables
TINYGOROOT := $(shell tinygo env TINYGOROOT)
WASM_EXEC := $(TINYGOROOT)/targets/wasm_exec.js
OUTPUT := bstates.wasm
BUILD_CMD := GOOS=js GOARCH=wasm tinygo build -o $(OUTPUT) -tags noasm --no-debug

# Default target
all: build link

# Compile the wasm file
build:
	@echo "Compiling the WASM file..."
	$(BUILD_CMD)
	@echo "File $(OUTPUT) generated."

# Create a symbolic link to wasm_exec.js
link:
	@echo "Creating a symbolic link for wasm_exec.js..."
	ln -sf $(WASM_EXEC) ./wasm_exec.js
	@echo "Symbolic link created: wasm_exec.js -> $(WASM_EXEC)"

# Clean generated files
clean:
	@echo "Cleaning up generated files..."
	rm -f $(OUTPUT) wasm_exec.js
	@echo "Files cleaned."
