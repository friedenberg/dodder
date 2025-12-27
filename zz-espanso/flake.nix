{
  description = "";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/d981d41ffe5b541eae3782029b93e2af5d229cc2";
    nixpkgs-stable.url = "github:NixOS/nixpkgs/09eb77e94fa25202af8f3e81ddc7353d9970ac1b";
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
