# Bstates from WASM

This module exports some functions in WASM to work with bstates from a Node.js backend or a browser with JavaScript.

## Build and test dependencies
Only [nix](https://nixos.org/) is required to generate and test the package. The [flakes](https://nixos.wiki/wiki/Flakes) experimental feature must be enabled in the nix installation.

## How to build
```
npm run build
```

## How to test
```
npm run test
```

## Examples

To execute the examples, you also need to create a shell with `nix develop` to ensure all dependencies (e.g., Node.js) are available to run the example programs.

All examples import and execute the function implemented in `example.js`

### Node.js ES
```bash
node examples/nodejs_es/main.js
```

### Node.js CommonJS
```bash
node examples/nodejs_cjs/main.js
```

### Web
The `examples/web` directory contains the files needed to serve a simple webpage where you can interact with the exported bstates functions through the browser console.

- From root directory, start the file server with:

```bash
./examples/web/server.sh
```

- Open [http://localhost:8080/examples/web](http://localhost:8080/examples/web) in a browser to experiment from the console.
