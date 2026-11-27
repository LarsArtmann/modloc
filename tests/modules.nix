# Eval checks for the NixOS and the Home Manager module, one pair per
# system. Each check evaluates the module enabled and disabled: enabled must
# install this flake's package exactly once, disabled must not install it. The
# host adds its own packages to the same list (NixOS its system tools), so
# the list is searched for the package, never compared whole. An option
# branch that is never evaluated is untested code.
{
  system,
  nixpkgs,
  flake,
}:

let
  inherit (nixpkgs) lib;
  pkgs = nixpkgs.legacyPackages.${system};
  pkg = flake.packages.${system}.default;

  systemPackages =
    enable:
    (lib.nixosSystem {
      inherit system;
      modules = [
        flake.nixosModules.modloc
        { programs.modloc.enable = enable; }
      ];
    }).config.environment.systemPackages;

  # Home Manager itself is not an input: the module is evaluated against a
  # stub that declares the one option it sets, home.packages.
  homeStub = {
    options.home.packages = lib.mkOption {
      type = lib.types.listOf lib.types.package;
      default = [ ];
    };
  };

  userPackages =
    enable:
    (lib.evalModules {
      specialArgs.pkgs = pkgs;
      modules = [
        homeStub
        flake.homeManagerModules.modloc
        { programs.modloc.enable = enable; }
      ];
    }).config.home.packages;

  check =
    name: packages:
    let
      enabled = packages true;
      disabled = packages false;
      # How often the package is in a list.
      installed = packages: lib.length (lib.filter (installed: installed == pkg) packages);
      ok = installed enabled == 1 && installed disabled == 0;
    in
    pkgs.runCommand "modloc-${name}-check" { } ''
      echo "enabled installs:  ${builtins.toJSON (map toString enabled)}" >&2
      echo "disabled installs: ${builtins.toJSON (map toString disabled)}" >&2
      echo "expected when enabled: ${pkg}" >&2
      ${if ok then "" else "echo \"modloc module check failed: ${name}\" >&2; exit 1"}
      touch $out
    '';
in
{
  nixos-module = check "nixos-module" systemPackages;
  home-manager-module = check "home-manager-module" userPackages;
}
