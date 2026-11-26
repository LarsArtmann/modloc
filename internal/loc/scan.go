package loc

import (
	"context"
	"fmt"
)

// Scan counts one repository: it finds the module markers, counts the files
// and attributes them. repo is both the directory that is walked and the
// label the report carries. An unknown language, an unreadable root or a
// failing counter is an error; nothing is counted partially.
func Scan(ctx context.Context, repo string, lang Language, counter FileCounter) (Report, error) {
	err := lang.validate()
	if err != nil {
		return Report{}, err
	}

	markers, skipped, err := FindMarkers(repo, lang)
	if err != nil {
		return Report{}, fmt.Errorf("find modules in %s: %w", repo, err)
	}

	files, err := counter.CountFiles(ctx, repo)
	if err != nil {
		return Report{}, fmt.Errorf("count files in %s: %w", repo, err)
	}

	return Report{
		Repo:    repo,
		Modules: Attribute(files, markers, lang),
		Skipped: skipped,
	}, nil
}
