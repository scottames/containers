// Fedora container image
// Generates a container image from the specified source Fedora Atomic image
// https://fedoraproject.org/atomic-desktops/
//
// TODO: support non-atomic
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// New initializes the Fedora Dagger module
func New(
	ctx context.Context,
	// Container registry
	// +optional
	// +default="quay.io"
	registry string,
	// Container registry organization
	// +optional
	// +default="fedora-ostree-desktops"
	org string,
	// Atomic variant
	// +optional
	// +default="silverblue"
	variant string,
	// Variant suffix string
	// e.g. main (as related to ublue-os images)
	// +optional
	suffix *string,
	// Tag or major release version
	// +optional
	// +default="40"
	tag string,
) *Fedora {
	return &Fedora{
		Tag:      tag,
		Org:      org,
		Registry: registry,
		Suffix:   suffix,
		Variant:  variant,
	}
}

// Fedora represents the constructed Fedora image
type Fedora struct {
	Org      string
	Registry string
	Suffix   *string
	Tag      string
	Variant  string
	Digests  []string

	Directories       []*DirectoryFromSource
	Files             []*FileFromSource
	PackagesInstalled []string
	PackagesRemoved   []string
	Repos             []*Repo
	ExecScriptPre     []*File
	ExecScriptPost    []*File
	ExecPre           [][]string
	ExecPost          [][]string
	Labels            []*ContainerLabel
}

// httpGet will get the given url and return the data
func (f *Fedora) httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting url '%s': %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status for url '%s': %v", url, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading data from url '%s': %w", url, err)
	}

	return data, nil
}
