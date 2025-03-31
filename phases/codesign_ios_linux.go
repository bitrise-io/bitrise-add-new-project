package phases

import (
	"errors"

	bitriseModels "github.com/bitrise-io/bitrise/v2/models"
)

func iosCodesign(bitriseYML bitriseModels.BitriseDataModel, searchDir string) (CodesignResultsIOS, error) {
	return CodesignResultsIOS{}, errors.New("Not supported on linux")
}
