{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/d981d41ffe5b541eae3782029b93e2af5d229cc2";
    nixpkgs-stable.url = "github:NixOS/nixpkgs/09eb77e94fa25202af8f3e81ddc7353d9970ac1b";
    utils.url = "https://flakehub.com/f/numtide/flake-utils/0.1.102";

    devenv-go.url = "github:friedenberg/eng?dir=pkgs/alfa/devenv-go";
    devenv-shell.url = "github:friedenberg/eng?dir=pkgs/alfa/devenv-shell";
  };

  outputs =
    { self
    , nixpkgs
    , nixpkgs-stable
    , utils
    , devenv-go
    , devenv-shell
    ,
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

        dodder = pkgs.buildGoApplication {
          pname = "dodder";
          version = "0.0.1";
          src = ./.;
          subPackages = [
            "cmd/der"
            # "cmd/dodder"
          ];
          modules = ./gomod2nix.toml;
        };

      in
      {

        packages.dodder = dodder;
        packages.default = dodder;

        docker = pkgs.dockerTools.buildImage {
          name = "dodder";
          tag = "latest";
          copyToRoot = [ dodder ];
          config = {
            Cmd = [ "${dodder}/bin/dodder" ];
            Env = [ ];
            ExposedPorts = { "9000/tcp" = { }; };
          };
        };

        devShells.default = pkgs.mkShell {
          # inherit (gomod2nix.packages.${system}) mkGoEnv gomod2nix;

          packages = (with pkgs; [
            bats
            fish
            gnumake
            gum
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
