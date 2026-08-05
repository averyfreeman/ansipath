{
  config,
  lib,
  pkgs,
  ...
}:

let
  cfg = config.programs.ansipath;
in
{
  options.programs.ansipath = {
    enable = lib.mkEnableOption "ansipath";

    package = lib.mkOption {
      type = lib.types.package;
      default = pkgs.callPackage ../../package.nix { };
      defaultText = lib.literalExpression "inputs.ansipath.packages.<system>.default";
      description = "The ansipath package to install.";
    };
  };

  config = lib.mkIf cfg.enable {
    home.packages = [ cfg.package ];
  };
}
