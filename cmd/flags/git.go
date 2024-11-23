package flags

import (
	"fmt"
)

type Git bool

const (
	Yes Git = true
	No  Git = false
)

var AllowedGitOptions = []Git{Yes, No}

func (g Git) String() string {
	if g {
		return "true"
	}
	return "false"
}

func (g *Git) Type() string {
	return "Git"
}

func (g *Git) Set(value string) error {
	boolValue := value == "true"
	for _, git := range AllowedGitOptions {
		if git == Git(boolValue) {
			*g = Git(boolValue)
			return nil
		}
	}

	return fmt.Errorf("Git flag. Allowed values: true, false")
}

// IsGitEnabled checks if the Git feature is enabled.
// It returns true if Git is enabled, otherwise false.
func (g *Git) IsGitEnabled() bool {
	return bool(*g)
}
