// Package loc attributes lines of code to the logical modules of a repository.
//
// A logical module is a directory that holds a module marker: go.mod for Go,
// Cargo.toml for Rust. Every counted file belongs to the nearest marker
// directory above it, so a nested module's files never count toward the
// module that encloses it. A file with no marker above it belongs to the
// repository root when the root is a module itself, and to the row OTHER when
// it is not.
package loc

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Language selects both halves of attribution: the marker file that declares a
// module and the file extension that makes a file count.
type Language string

const (
	// Go attributes by go.mod and counts .go files.
	Go Language = "go"
	// Rust attributes by Cargo.toml and counts .rs files.
	Rust Language = "rust"
)

// ErrUnknownLanguage rejects a value that names no language.
var ErrUnknownLanguage = errors.New("unknown language")

// AllLanguages lists every language, in the order the --language help shows.
func AllLanguages() []Language {
	return []Language{Go, Rust}
}

// Parse converts s to a Language, or returns an error wrapping ErrUnknownLanguage.
func Parse(s string) (Language, error) {
	if slices.Contains(AllLanguages(), Language(s)) {
		return Language(s), nil
	}

	return "", fmt.Errorf("%w: %q (allowed: %s)", ErrUnknownLanguage, s, joinNames(AllLanguages()))
}

// String returns the language as typed on the command line.
func (l Language) String() string {
	return string(l)
}

// Marker is the file that declares a module in this language.
func (l Language) Marker() string {
	switch l {
	case Go:
		return "go.mod"
	case Rust:
		return "Cargo.toml"
	default:
		return ""
	}
}

// Extension is the file extension counted in this language, dot included.
func (l Language) Extension() string {
	switch l {
	case Go:
		return ".go"
	case Rust:
		return ".rs"
	default:
		return ""
	}
}

// Others lists every language except l, for hints that point at a re-run.
func (l Language) Others() []Language {
	return slices.DeleteFunc(slices.Clone(AllLanguages()), func(other Language) bool { return other == l })
}

// validate rejects a Language that was not built by Parse or a constant.
func (l Language) validate() error {
	if slices.Contains(AllLanguages(), l) {
		return nil
	}

	return fmt.Errorf("%w: %q (allowed: %s)", ErrUnknownLanguage, l, joinNames(AllLanguages()))
}

// joinNames renders languages the way they are typed on the command line.
func joinNames(languages []Language) string {
	names := make([]string, 0, len(languages))
	for _, language := range languages {
		names = append(names, language.String())
	}

	return strings.Join(names, ", ")
}
