package workflow

import "fmt"

// DefaultImageMap maps common GitHub Actions runner labels to Docker images.
var DefaultImageMap = map[string]string{
	"ubuntu-latest":  "ghcr.io/catthehacker/ubuntu:act-latest",
	"ubuntu-24.04":   "ghcr.io/catthehacker/ubuntu:act-24.04",
	"ubuntu-22.04":   "ghcr.io/catthehacker/ubuntu:act-22.04",
	"ubuntu-20.04":   "ghcr.io/catthehacker/ubuntu:act-20.04",
	"ubuntu-18.04":   "ghcr.io/catthehacker/ubuntu:act-18.04",
}

// ResolveImage resolves a runs-on label to a Docker image.
// Platform overrides take priority over defaults.
func ResolveImage(runsOn []string, overrides map[string]string) (string, error) {
	for _, label := range runsOn {
		if img, ok := overrides[label]; ok {
			return img, nil
		}
		if img, ok := DefaultImageMap[label]; ok {
			return img, nil
		}
	}
	// Fallback: if the label looks like a Docker image (contains /), use it directly
	for _, label := range runsOn {
		if len(label) > 0 && (containsSlash(label) || containsColon(label)) {
			return label, nil
		}
	}
	if len(runsOn) > 0 {
		return "", fmt.Errorf("unsupported runner %q — add a platform override with --platform %s=<image>", runsOn[0], runsOn[0])
	}
	return "", fmt.Errorf("runs-on is empty")
}

func containsSlash(s string) bool {
	for _, c := range s {
		if c == '/' {
			return true
		}
	}
	return false
}

func containsColon(s string) bool {
	for _, c := range s {
		if c == ':' {
			return true
		}
	}
	return false
}
