package main

import "context"

type DirectoryFromSource struct {
	Source      *Directory
	Destination string
}

// TODO: docs
func (f *Fedora) WithDirectory(ctx context.Context, destination string, source *Directory) *Fedora {
	dir := DirectoryFromSource{Source: source, Destination: destination}
	f.Directories = append(f.Directories, &dir)

	return f
}
