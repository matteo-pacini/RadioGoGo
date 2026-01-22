{
  description = "RadioGoGo Flake";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {
        inherit system;
        overlays = [
          (final: prev: {
            radiogogo = prev.callPackage ./nix/package.nix {};
          })
        ];
      };
    in {
      devShells.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          go
          delve
          gopls
          go-tools
          gotools
          ffmpeg
          zip
        ];
      };
      packages = {
        radiogogo = pkgs.radiogogo;
      };
    });
}
