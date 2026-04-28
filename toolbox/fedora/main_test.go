package main

import "testing"

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
