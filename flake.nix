{
  description = "ansipath: a colorized diagnostic view of shell PATH entries";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-26.05-darwin";

    nix-darwin = {
      url = "github:nix-darwin/nix-darwin/nix-darwin-26.05";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    home-manager = {
      url = "github:nix-community/home-manager/release-26.05";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    { self, nixpkgs, ... }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
      packageFor =
        system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        pkgs.callPackage ./package.nix { };
    in
    {
      packages = forAllSystems (system: {
        ansipath = packageFor system;
        default = self.packages.${system}.ansipath;
      });

      nixosModules.default = import ./nix/modules/nixos.nix;
      darwinModules.default = import ./nix/modules/darwin.nix;
      homeManagerModules.default = import ./nix/modules/home-manager.nix;

      formatter = forAllSystems (system: (import nixpkgs { inherit system; }).nixfmt-rfc-style);
    };
}
