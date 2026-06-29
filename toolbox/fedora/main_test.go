package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestDaggerConfigDoesNotInstallDistroboxHelpers(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("dagger.json")
	if err != nil {
		t.Fatalf("read dagger.json: %v", err)
	}

	var config struct {
		Dependencies []struct {
			Name string `json:"name"`
		} `json:"dependencies"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("parse dagger.json: %v", err)
	}

	for _, dep := range config.Dependencies {
		if dep.Name == "distrobox" {
			t.Fatal("toolbox image should rely on distrobox create-time integration, not the daggerverse distrobox helper")
		}
	}
}

func TestHasCompatibleMesaFreeworldDrivers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		releaseVersion string
		want           bool
	}{
		{
			name:           "fedora 43 keeps freeworld mesa drivers",
			releaseVersion: "43",
			want:           true,
		},
		{
			name:           "fedora 44 skips freeworld mesa drivers",
			releaseVersion: "44",
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := hasCompatibleMesaFreeworldDrivers(tt.releaseVersion)
			if got != tt.want {
				t.Fatalf("hasCompatibleMesaFreeworldDrivers(%q) = %t, want %t", tt.releaseVersion, got, tt.want)
			}
		})
	}
}
