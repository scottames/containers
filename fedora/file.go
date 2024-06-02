package main

import "context"

type FileFromSource struct {
	Destination string
	Source      *File
}

// TODO: docs
func (f *Fedora) WithFile(ctx context.Context, destination string, source *File) *Fedora {
	file := FileFromSource{Source: source, Destination: destination}
	f.Files = append(f.Files, &file)

	return f
}
