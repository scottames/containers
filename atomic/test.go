package main

import (
	"context"
	"fmt"
)

// FIXME: !no commit
func (a *Atomic) Test(ctx context.Context) (string, error) {
	return dag.Module().Name(ctx)
}

// FIXME: !no commit
func (a *Atomic) PrintContainerVersion(ctx context.Context) ([]string, error) {
	_, err := a.fedoraAtomic(ctx)
	if err != nil {
		return nil, err
	}
	return []string{
		fmt.Sprintf("---"),
		fmt.Sprintf("base_image:         %s", a.BaseImage),
		fmt.Sprintf("base_image_version: %s", a.BaseImageVersion),
		fmt.Sprintf("---"),
		fmt.Sprintf("major_version:      %s", *a.MajorVersion),
		fmt.Sprintf("timestamp:          %s", a.Date),
	}, nil
}
