# bstates with WebAssembly (WASM)

This module exports functions implemented in WebAssembly (WASM), allowing you to work with **bstates** both from a Node.js backend and directly in a browser using JavaScript.

---

## Build and Test Dependencies

To build and test this project, you will need:

1. [**Nix**](https://nixos.org/download/) (with the flakes experimental feature enabled).  
   - **Why Nix?**  
     Nix ensures the correct version of the Go compiler is available to generate the `bstates.wasm` file and guarantees tests run with a compatible version of Node.js.  
   - **See also:** [Enabling flakes in Nix](#enabling-flakes-in-nix).

2. **Node.js**  
   - If Node.js is not installed on your system, you can still use the [Nix shell](#running-a-nix-shell) to access it.

3. **npm**  
   - Similarly, if `npm` is not installed locally, you can use the Nix shell environment to run it.

---

### Enabling Flakes in Nix

If [flakes](https://nixos.wiki/wiki/Flakes) are not already enabled on your system, add the following lines to your Nix configuration:

```bash
mkdir -p ~/.config/nix
echo 'experimental-features = nix-command flakes' >> ~/.config/nix/nix.conf
```

### Running a Nix Shell
From the root of the project (or any subdirectory), run:

```bash
nix develop
```
This command opens a Nix shell, providing all the tools and versions required to build and test the project.

## How to Build and Install Dependencies

To build the project and install all dependencies, run:
```bash
npm install
```
This command will:
1. Automatically open a Nix shell to ensure the correct environment is set up (including the Go compiler).
2. Generate the bstates.wasm file.
3. Install all required JavaScript dependencies.

## How to Test

To run the test suite, execute:

```bash
npm test
```
This command will:
1. Automatically open a Nix shell to ensure a consistent environment.
2. Run all tests using Jest.

## Examples

The function defined in example.js is executed in each example. All examples follow these general steps:
1. Navigate to the example directory in `examples` directory (e.g., examples/nodejs_es).
2. Install dependencies:
```bash
npm install
```
3. Start the example:
```bash
npm start
```
### Node.js ES Module (ESM)

- **Directory**: `nodejs_es`
- **Overview**: Demonstrates loading `bstates.wasm` and using it in an ES Module environment.  
  After running `npm start`, the code will execute in a Node.js ESM context.

### Node.js CommonJS (CJS)

- **Directory**: `nodejs_cjs`
- **Overview**: Demonstrates loading `bstates.wasm` and using it in a CommonJS environment.  
  After running `npm start`, the code will execute in a Node.js CJS context.

### Web Client (Fetching the `.wasm` File)

- **Directory**: `examples/web_raw`  
- **Overview**: This setup downloads the `.wasm` file directly from a specified directory (or URL).  
- **How It Works**: You provide a path when calling the load function, and the WASM file is fetched over HTTP just like any other resource. This method keeps the `.wasm` separate from your JavaScript code, improving caching efficiency and allowing for streaming in modern browsers.

### Web Client (Base64-Encoded WASM)

- **Directory**: `examples/web_b64`  
- **Overview**: In this approach, the `.wasm` file is embedded in a JavaScript module as Base64-encoded data.  
- **How It Works**: No path or separate file is needed; the WASM bytes are bundled within the JavaScript. This eliminates external file hosting requirements but results in a slightly larger download and an additional decoding step.

Both examples automatically execute the sample function once the WASM is loaded, and they expose the resulting object in the browser console for interactive experimentation.  

### Additional Notes
- **API Documentation**: Refer to the inline documentation in `example.js` for details on using the `bstates` API.  
- **Troubleshooting**:  
  - Ensure that Node.js and npm are installed, or use the Nix shell.  
  - Verify all Go and JavaScript dependencies are properly installed.  
- Feel free to open an issue or reach out if you have any questions or need further support!
