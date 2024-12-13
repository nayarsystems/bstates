# Take a look to doc
https://tinygo.org/docs/guides/webassembly/wasm/
https://github.com/tinygo-org/tinygo/tree/release/src/examples/wasm


# Build using tinygo 
GOOS=js GOARCH=wasm tinygo build -o wasm.wasm -tags noasm --no-debug

# Get tinygo wasm importer (wasm_exec.js)
tinygoRoot=$(tinygo env TINYGOROOT)

cp ${tinygoRoot}/targets/wasm_exec.js wasm_exec.js

# Serve current dir
goexec 'http.ListenAndServe(`:8080`, http.FileServer(http.Dir(`.`)))'
