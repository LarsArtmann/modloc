package loc

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
)

// Pruned directories are never descended when looking for markers. They hold
// third-party or build output, not modules of the repository: vendor and
// node_modules are dependency trees, target is Cargo's build directory, and
// .git is the repository's own metadata.
var prunedDirs = []string{"vendor", "target", "node_modules", ".git"}

// FindMarkers lists every directory under root that holds a marker file of the
// language, relative to root and slash-separated; the root itself is
// RootModule. Pruned directories are skipped. A directory that cannot be read
// is returned in skipped and its subtree is not searched, so the result
// covers the readable tree only. An unreadable root is an error: nothing can
// be attributed without it.
func FindMarkers(root string, lang Language) (markers []string, skipped []string, err error) {
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		switch {
		case walkErr != nil && path == root:
			return walkErr
		case walkErr != nil:
			skipped = append(skipped, relative(root, path))

			return nil
		case !entry.IsDir():
			if entry.Name() == lang.Marker() {
				markers = append(markers, relative(root, filepath.Dir(path)))
			}

			return nil
		case isPruned(root, path, entry):
			return fs.SkipDir
		default:
			return nil
		}
	})
	if err != nil {
		return nil, nil, fmt.Errorf("walk %s: %w", root, err)
	}

	return markers, skipped, nil
}

// isPruned reports whether a directory is one of the pruned names. The root
// is never pruned: a repository may itself be called vendor or target.
func isPruned(root, path string, entry fs.DirEntry) bool {
	return path != root && slices.Contains(prunedDirs, entry.Name())
}

// relative expresses path under root the way report rows are written.
func relative(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return Normalize(path)
	}

	return Normalize(rel)
}
