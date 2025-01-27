{
  description = "Nix environment to generate the bstates wasm file and execute wasm examples";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs";
  };

  outputs = { self, nixpkgs }: 
  let
    pkgs = nixpkgs.legacyPackages.x86_64-linux;
  in {
    devShells.x86_64-linux = {
        default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go_1_22
            tinygo
            nodejs_23
          ];
        shellHook = ''
          echo "======================================="
          echo "Build the wasm file from wasm directory by running \"make\"."
          echo "Ensure that all npm dependencies are installed by running \"npm install\"."
          echo ""
          echo "The \"examples\" directory contains examples of how to use the bstates from Node.js and Browser." 
          echo "To run an example go to the directory of an example and run \"npm start\"."
          echo "Before running each example, you need to run \"npm install\" from the directory of the example."
          echo "Examples:"
          echo "- nodejs_cjs: Node.js program that follows the \"CommonJS\" style."
          echo "- nodejs_es: Node.js program that follows the \"ES modules\" style."
          echo "- web: client side program from browser."
          echo "======================================="
        '';
        };
    };
  };
}
