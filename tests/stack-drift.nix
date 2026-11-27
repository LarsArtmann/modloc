# STACK.md drift check. STACK.md is the decision record of the foundation
# (the stack-selection skill writes it); without a reader it is a ghost
# system. This is the reader, for the one module of this repository: every
# first-party module that go.mod requires directly must be named in the
# "Chosen foundation" section of STACK.md, by the first segment of its path
# (`github.com/larsartmann/go-output/table` is `go-output`). Indirect
# requires are decided by the modules themselves, and third-party
# dependencies are how-to-golang's domain.
#
# Drift fails the evaluation, before anything is built.
{
  lib,
  pkgs,
  # go.mod and STACK.md, as store paths.
  goMod,
  stack,
}:

let
  indirect = "// indirect";

  # section is the text under one "## " heading, up to the next one.
  section =
    name: text:
    let
      parts = lib.split "\n## ${name}\n" ("\n" + text);
    in
    if builtins.length parts < 3 then
      throw "modloc stack drift: STACK.md has no '## ${name}' section"
    else
      builtins.head (lib.split "\n## " (lib.elemAt parts 2));

  stackText = builtins.readFile stack;
  chosen = section "Chosen foundation" stackText;
  excluded = section "Deliberately excluded" stackText;

  # library is the first path segment of a first-party module a go.mod line
  # requires directly, or null for every other line. A line of the require
  # block and a one-line require both read "github.com/… <version>" once the
  # keyword and the blanks are gone.
  library =
    line:
    let
      match = builtins.match "github\\.com/larsartmann/([^[:space:]/]+)[^[:space:]]*[[:space:]]+v[^[:space:]]+" (
        lib.trim (lib.removePrefix "require " (lib.trim line))
      );
    in
    if lib.hasInfix indirect line || match == null then null else builtins.head match;

  libraries = lib.naturalSort (
    lib.unique (
      lib.filter (name: name != null) (map library (lib.splitString "\n" (builtins.readFile goMod)))
    )
  );

  undeclared = lib.filter (name: !lib.hasInfix name chosen) libraries;
  contradictions = lib.filter (name: lib.hasInfix name excluded) undeclared;
  nowhere = lib.filter (name: !lib.hasInfix name excluded) undeclared;

  findings =
    lib.optional (contradictions != [ ])
      "required by go.mod but listed as deliberately excluded: ${lib.concatStringsSep ", " contradictions}"
    ++
      lib.optional (nowhere != [ ])
        "required by go.mod but named nowhere in the Chosen foundation section: ${lib.concatStringsSep ", " nowhere}";
in
lib.throwIfNot (findings == [ ]) "modloc stack drift: ${lib.concatStringsSep "; " findings}" (
  pkgs.runCommand "modloc-stack-drift-check" { } ''
    echo "modloc stack drift: ${lib.concatStringsSep ", " libraries} declared in STACK.md" >&2
    touch $out
  ''
)
