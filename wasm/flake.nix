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
          echo 
          echo "To test the NodeJs example run  \"node nodejs/example/main.js\"."
          echo 
          echo "To test the web example go to web directory and:"
          echo "- run a web server: ./server.sh"
          echo "- open a browser and go to http://localhost:8080"
          echo "- open debug console and call bstates exported functions"
          echo "======================================="
        '';
        };
    };
  };
}
