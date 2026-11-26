# NixOS module: put modloc on the system PATH.
#
# Usage:
#   imports = [ inputs.modloc.nixosModules.modloc ];
#   programs.modloc.enable = true;
#
# programs.modloc is the option path the bash modloc of the nixos-modules
# collection declared, so a consumer that moves over changes its import and
# nothing else. The package defaults to this flake's own build, which wraps
# scc onto the program's PATH; override it to pin or re-wrap.
{
  config,
  lib,
  pkgs,
  modlocFlake,
  ...
}:

let
  cfg = config.programs.modloc;
in
{
  options.programs.modloc = {
    enable = lib.mkEnableOption "modloc, the per-module lines-of-code counter, on the system PATH";

    package = lib.mkOption {
      type = lib.types.package;
      default = modlocFlake.packages.${pkgs.stdenv.hostPlatform.system}.default;
      defaultText = lib.literalExpression "modloc.packages.\${pkgs.stdenv.hostPlatform.system}.default";
      description = "The modloc package to put on PATH.";
    };
  };

  config = lib.mkIf cfg.enable {
    environment.systemPackages = [ cfg.package ];
  };
}
