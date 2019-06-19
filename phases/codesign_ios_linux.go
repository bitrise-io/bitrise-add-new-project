package phases

import bitriseModels "github.com/bitrise-io/bitrise/models"
import "errors"

func iosCodesign(bitriseYML bitriseModels.BitriseDataModel, searchDir string) (CodesignResultsIOS, error) {
	return CodesignResultsIOS{}, errors.New("Not supported on linux")
}
