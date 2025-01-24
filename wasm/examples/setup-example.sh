#!/bin/sh

# Get the current directory (where the script is executed)
EXAMPLE_DIR=$(pwd)

# Get the project root directory (parent directory of 'examples')
PROJECT_ROOT=$(cd "$(dirname "$0")/.." && pwd)

# Path to the node_modules directory of the example
NODE_MODULES_DIR="$EXAMPLE_DIR/node_modules"

# Path to the bstates package in the project root
BSTATES_PACKAGE_DIR="$PROJECT_ROOT"

# Path to the symlink that needs to be created
LINK_PATH="$NODE_MODULES_DIR/bstates"

# Create the node_modules directory if it does not exist
if [ ! -d "$NODE_MODULES_DIR" ]; then
    echo "Creating node_modules directory in $EXAMPLE_DIR"
    mkdir -p "$NODE_MODULES_DIR"
fi

# Create the symlink if it does not exist
if [ ! -L "$LINK_PATH" ]; then
    echo "Creating symlink for bstates in $NODE_MODULES_DIR"
    ln -s "$BSTATES_PACKAGE_DIR" "$LINK_PATH"
    echo "Symlink created: $LINK_PATH -> $BSTATES_PACKAGE_DIR"
fi
