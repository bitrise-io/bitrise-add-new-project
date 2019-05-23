package phases

import (
	"testing"

	"github.com/bitrise-io/bitrise/models"
)

func TestGetDefaultStack(t *testing.T) {
	if stack, _, err := getProjectInfo(models.BitriseDataModel{
		FormatVersion:        "7",
		DefaultStepLibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
		ProjectType:          "android",
	}); err != nil {
		t.Fatalf("get default stack from recognized bitrise.yml: %s", err)
	} else if stack == "" {
		t.Fatalf("get default stack from recognized bitrise.yml shoulde be android")
	}

	if stack, _, err := getProjectInfo(models.BitriseDataModel{
		FormatVersion:        "7",
		DefaultStepLibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
	}); err != nil {
		t.Fatalf("get default stack from unrecognized bitrise.yml should not be error")
	} else if stack != "" {
		t.Fatalf("get default stack from unrecognized bitrise.yml shoulde be empty")
	}
}
