package main

// Fedora 44 and 45 RPM Fusion mesa-freeworld packages can lag Fedora mesa versions,
// which makes the swap transaction fail until RPM Fusion catches up.
func hasCompatibleMesaFreeworldDrivers(releaseVersion string) bool {
	return releaseVersion != "44" && releaseVersion != "45"
}
