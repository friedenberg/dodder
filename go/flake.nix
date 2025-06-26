{
  inputs = {
    nixpkgs-stable.url = "https://flakehub.com/f/NixOS/nixpkgs/0.2411.717296.tar.gz";
    utils.url = "https://flakehub.com/f/numtide/flake-utils/0.1.102.tar.gz";

    devenv-go.url = "github:friedenberg/eng?dir=pkgs/alfa/devenv-go";
    devenv-shell.url = "github:friedenberg/eng?dir=pkgs/alfa/devenv-shell";
  };

  outputs = {
    self,
    nixpkgs,
    nixpkgs-stable,
    utils,
    devenv-go,
    devenv-shell,
  }:
    (utils.lib.eachDefaultSystem
      (system:
        let

          pkgs = import nixpkgs {
            inherit system;

            overlays = [
              devenv-go.overlays.default
            ];
          };

          zit = pkgs.buildGoApplication {
            pname = "zit";
            version = "0.0.1";
            src = ./.;
            modules = ./gomod2nix.toml;
          };

        in {

          packages.zit = zit;
          packages.default = zit;

          docker = pkgs.dockerTools.buildImage {
            name = "zit";
            tag = "latest";
            copyToRoot = [zit];
            config = {
              Cmd = ["${zit}/bin/zit"];
              Env = [];
              ExposedPorts = {"9000/tcp" = {};};
            };
          };

          devShells.default = pkgs.mkShell {
            # inherit (gomod2nix.packages.${system}) mkGoEnv gomod2nix;

            packages = (with pkgs; [
              govulncheck
              bats
              fish
              gnumake
              just
            ]);

            inputsFrom = [
              devenv-go.devShells.${system}.default
              devenv-shell.devShells.${system}.default
            ];
          };
        })
    );
}
