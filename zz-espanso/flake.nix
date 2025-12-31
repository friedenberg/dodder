{
  description = "";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/c4cfc9ced33f81099f419fa59893df11dc3f9de9";
    nixpkgs-stable.url = "github:NixOS/nixpkgs/9ef261221d1e72399f2036786498d78c38185c46";
    utils.url = "https://flakehub.com/f/numtide/flake-utils/0.1.102";
  };

  outputs = { self, nixpkgs, nixpkgs-stable, utils }:
    utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        name = "generate";
        buildInputs = with pkgs; [ jq yq-go just ];
        generate = (
          pkgs.writeScriptBin name (builtins.readFile ./generate.bash)
        ).overrideAttrs(old: {
          buildCommand = "${old.buildCommand}\n patchShebangs $out";
        });
      in rec {
        packages.generate = pkgs.symlinkJoin {
          name = name;
          paths = [ generate ] ++ buildInputs;
          buildInputs = [ pkgs.makeWrapper ];
          postBuild = "wrapProgram $out/bin/${name} --prefix PATH : $out/bin";
        };

        defaultPackage = packages.generate;

        devShells.default = pkgs.mkShell {
          packages = (with pkgs; [
            jq
            yq-go
            just
          ]);

          inputsFrom = [];
        };
      }
    );
}
