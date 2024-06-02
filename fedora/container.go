package main

import (
	"context"
	"fmt"
)

// TODO: docs
func (f *Fedora) Container(ctx context.Context) (*Container, error) {
	suffix := ""
	if f.Suffix != nil {
		suffix = fmt.Sprintf("-%s", *f.Suffix)
	}

	const sourceStr = "%s/%s/%s%s:%s" // registry/org/variant+suffix:tag
	ctr, err := f.
		ContainerFrom(
			ctx,
			fmt.Sprintf(sourceStr,
				f.Registry,
				f.Org,
				f.Variant,
				suffix,
				f.Tag,
			),
		)
	if err != nil {
		return nil, err
	}

	return ctr, nil
}

func (f *Fedora) ContainerFrom(
	ctx context.Context,
	// TODO: docs
	from string,
) (*Container, error) {
	ctr := dag.
		Container().
		From(from)

	ctr = f.ctrWithDirectoriesInstalled(ctr)
	ctr = f.ctrWithFilesInstalled(ctr)

	if f.Repos != nil {
		var err error
		ctr, err = f.ctrWithReposInstalled(ctr)
		if err != nil {
			return nil, err
		}
	}

	if f.ExecScriptPre != nil {
		var err error
		ctr, err = f.ctrExecScripts(ctx, ctr, f.ExecScriptPre)
		if err != nil {
			return nil, err
		}
	}

	for _, cmd := range f.ExecPre {
		ctr = ctr.WithExec(cmd)
	}

	if f.PackagesInstalled != nil || f.PackagesRemoved != nil {
		ctr = f.ctrWithPackagesInstalledAndRemoved(ctr)
	}

	ctr = f.ctrWithReposRemoved(ctr)

	if f.ExecScriptPost != nil {
		var err error
		ctr, err = f.ctrExecScripts(ctx, ctr, f.ExecScriptPost)
		if err != nil {
			return nil, err
		}
	}

	if f.ExecScriptPre != nil || f.ExecScriptPost != nil {
		scripts := append(f.ExecScriptPre, f.ExecScriptPost...)
		var err error
		ctr, err = f.ctrScriptsCleanup(ctx, ctr, scripts)
		if err != nil {
			return nil, err
		}
	}

	for _, cmd := range f.ExecPost {
		ctr = ctr.WithExec(cmd)
	}

	ctr = f.ctrWithLabels(ctr)

	return ctr, nil
}

func (f *Fedora) ctrWithDirectoriesInstalled(ctr *Container) *Container {
	for _, d := range f.Directories {
		ctr = ctr.WithDirectory(d.Destination, d.Source)
	}

	return ctr
}

func (f *Fedora) ctrWithFilesInstalled(ctr *Container) *Container {
	for _, d := range f.Files {
		ctr = ctr.WithFile(d.Destination, d.Source)
	}

	return ctr
}

func (f *Fedora) ctrWithLabels(ctr *Container) *Container {
	for _, l := range f.Labels {
		ctr = ctr.WithLabel(l.Name, l.Value)
	}

	return ctr
}

func (f *Fedora) ctrWithExec(ctr *Container, exec []string, args ...string) *Container {
	if args != nil {
		exec = append(exec, args...)
	}

	return ctr.WithExec(exec)
}

func (f *Fedora) ctrExecScripts(
	ctx context.Context,
	ctr *Container,
	scripts []*File,
) (*Container, error) {
	for _, script := range scripts {
		scriptName, err := script.Name(ctx)
		if err != nil {
			return nil, err
		}
		scriptTmp := fmt.Sprintf("/tmp/%s", scriptName)
		ctr = ctr.WithFile(scriptTmp, script).WithExec([]string{scriptTmp})
	}

	return ctr, nil
}

func (f *Fedora) ctrScriptsCleanup(
	ctx context.Context,
	ctr *Container,
	scripts []*File,
) (*Container, error) {
	filesToDelete := []string{}
	for _, script := range scripts {
		scriptName, err := script.Name(ctx)
		if err != nil {
			return nil, err
		}
		scriptTmp := fmt.Sprintf("/tmp/%s", scriptName)
		filesToDelete = append(filesToDelete, scriptTmp)
	}
	cmd := append([]string{"rm", "-f"}, filesToDelete...)
	return ctr.WithExec(cmd), nil
}
