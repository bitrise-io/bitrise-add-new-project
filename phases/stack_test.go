package phases

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDefaultStack(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {

	}

	if stack, _, err := getProjectInfo(filepath.Join(cwd, "..", "test", "known-project-type-bitrise.yml")); err != nil {
		t.Fatalf("get default stack from recognized bitrise.yml: %s", err)
	} else if stack == "" {
		t.Fatalf("get default stack from recognized bitrise.yml shoulde be android")
	}

	if stack, _, err := getProjectInfo(filepath.Join(cwd, "..", "test", "unknown-project-type-bitrise.yml")); err != nil {
		t.Fatalf("get default stack from unrecognized bitrise.yml should not be error")
	} else if stack != "" {
		t.Fatalf("get default stack from unrecognized bitrise.yml shoulde be empty")
	}
}
