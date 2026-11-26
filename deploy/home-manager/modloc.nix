# Home Manager module: put modloc on the user's PATH.
#
# Usage:
#   imports = [ inputs.modloc.homeManagerModules.modloc ];
#   programs.modloc.enable = true;
#
# Same option path as the NixOS module, for a user profile instead of a
# system: a machine that is not NixOS, or one user among several. The package
# defaults to this flake's own build, which wraps scc onto the program's
# PATH; override it to pin or re-wrap.
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
    enable = lib.mkEnableOption "modloc, the per-module lines-of-code counter, on the user's PATH";

    package = lib.mkOption {
      type = lib.types.package;
      default = modlocFlake.packages.${pkgs.stdenv.hostPlatform.system}.default;
      defaultText = lib.literalExpression "modloc.packages.\${pkgs.stdenv.hostPlatform.system}.default";
      description = "The modloc package to put on PATH.";
    };
  };

  config = lib.mkIf cfg.enable {
    home.packages = [ cfg.package ];
  };
}
