# Bstates from WASM

This module exports some functions in WASM to work with bstates from a Node.js backend or a browser with JavaScript.

## Compilation

First, it is necessary to generate the `bstates.wasm` file. To do this, from the root directory of the module, you need to create a Nix environment with the required dependencies. Therefore, this step requires having Nix installed.

To get a prepared shell with the compilation environment, run the following Nix command:

```bash
nix develop
```

The first time you run the command, it may take a while since it needs to download and cache all the required dependencies for the binary generation.

Once the command finishes, you can generate the binary by running:

```bash
make
```

## Examples

To execute the examples, you also need to create a shell with `nix develop` to ensure all dependencies (e.g., Node.js) are available to run the example programs.

### Node.js

The `nodejs` directory contains a Node.js module and sample program in Node.js, which can be executed with:

```bash
node nodejs/example/main.js
```

### Web

The `web` directory contains the files needed to serve a simple webpage where you can interact with the exported bstates functions through the browser console.

- From that directory, start the file server with:

```bash
./server.sh
```

- Open [http://localhost:8080/](http://localhost:8080/) in a browser to experiment from the console.
