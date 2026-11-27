{
  description = "modloc — lines of code per logical module (nearest go.mod or Cargo.toml), not per repository";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    go-nix-helpers = {
      url = "github:LarsArtmann/go-nix-helpers";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{
      self,
      flake-parts,
      nixpkgs,
      ...
    }:
    let
      inherit (nixpkgs) lib;
      packageSrc = lib.fileset.toSource {
        root = ./.;
        fileset = lib.fileset.gitTracked ./.;
      };
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
      # What --version prints. A flake sees no tags, so a clean checkout is
      # named by its short commit; a working tree is "dev" plus that commit.
      version = self.shortRev or "dev";
      commit = self.dirtyShortRev or "";
      # The checks over the whole repository, one set per system.
      repositoryChecks =
        system:
        (import ./tests/modules.nix {
          inherit system;
          inherit (inputs) nixpkgs;
          flake = self;
        })
        // {
          # STACK.md must name every first-party module that go.mod requires.
          stack-drift = import ./tests/stack-drift.nix {
            inherit (nixpkgs) lib;
            pkgs = nixpkgs.legacyPackages.${system};
            goMod = ./go.mod;
            stack = ./STACK.md;
          };

          # Both remotes are public and a daemon commits whatever lies around,
          # so every tracked file is scanned for secrets. --no-git: the store
          # copy carries no history.
          secrets = nixpkgs.legacyPackages.${system}.runCommand "modloc-secrets-check" {
            nativeBuildInputs = [ nixpkgs.legacyPackages.${system}.gitleaks ];
          } ''
            gitleaks detect --no-git --source ${packageSrc} --redact --no-banner >gitleaks.log 2>&1 && touch $out || {
              cat gitleaks.log >&2
              echo "modloc secrets check failed: gitleaks found a secret in the source tree" >&2
              exit 1
            }
          '';
        };
    in
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [ inputs.go-nix-helpers.flakeModules.go-standard ];

      go-standard = {
        pname = "modloc";
        description = "Lines of code per logical module (nearest go.mod or Cargo.toml), not per repository";
        goPkgAttr = "go_1_27";
        inherit systems;
        src = packageSrc;
        subPackages = [ "cmd/modloc" ];
        vendorHash = "sha256-tOVZYhI3tiSBPm8k0ARaimrdTAaKDh12XEEmlYaSQYE=";
        enableCompletions = true;
        completionStyle = "subcommand";
        lintAsCheck = true;
        ldflags = [
          "-s"
          "-w"
          "-X main.version=${version}"
          "-X main.commit=${commit}"
        ];
        # modloc counts through the scc binary. The wrapper puts that scc
        # on the program's PATH, so the package works on any machine; the
        # check phase gets the same scc for the specs that run the real one.
        extraBuildAttrs = pkgs: {
          nativeBuildInputs = [ pkgs.makeWrapper ];
          checkInputs = [ pkgs.scc ];
          postInstall = ''
            wrapProgram $out/bin/modloc --prefix PATH : ${lib.makeBinPath [ pkgs.scc ]}
          '';
        };
      };

      flake = {
        # programs.modloc: the same option path the bash modloc of the
        # nixos-modules collection declared, so a consumer switches with the
        # import alone. The package defaults to this flake's own build.
        nixosModules.modloc = {
          imports = [ ./deploy/nixos/modloc.nix ];
          _module.args.modlocFlake = self;
        };

        homeManagerModules.modloc = {
          imports = [ ./deploy/home-manager/modloc.nix ];
          _module.args.modlocFlake = self;
        };

        checks = builtins.listToAttrs (
          map (system: {
            name = system;
            value = repositoryChecks system;
          }) systems
        );
      };

      perSystem =
        { self', system, ... }:
        {
          # Runs the package the way a host does. This derivation has no scc,
          # so counting works only through the wrapper the package carries.
          checks.package-runtime = nixpkgs.legacyPackages.${system}.runCommand "modloc-package-runtime" { } ''
            bash ${./tests/package-runtime.sh} ${lib.escapeShellArgs [
              self'.packages.default
              version
              commit
            ]}
            touch $out
          '';
        };
    };
}
