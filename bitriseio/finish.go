package bitriseio

import (
	"fmt"
	"net/http"
)

// configByProjectType maps default config names to project types.
// Register finish endpoint is prepared for the use case of the bitrise.io website's frontend,
// where the user can let the scanner to generate a scan result or use 'manual' config,
// in which case the user selects one of our static default configs.
// Since in case of local project registration the frontend is not involved, we can use only the 'manual' config
// and select any of the give project type's default configs.
// Later the tool updated the project's bitrise.yml by calling the '/apps/slug/bitrise.yml' endpoint.
var configByProjectType = map[string]string{
	"android":      "default-android-config",
	"cordova":      "default-cordova-config",
	"fastlane":     "default-fastlane-android-config",
	"flutter":      "flutter-config-app-android",
	"ionic":        "default-ionic-config",
	"ios":          "default-ios-config",
	"macos":        "default-macos-config",
	"react-native": "default-react-native-config",
	"xamarin":      "default-xamarin-config",
	"other":        "other-config",
}

// RegisterFinishParams ...
type RegisterFinishParams struct {
	OrganizationSlug string         `json:"organization_slug"`
	ProjectType      string         `json:"project_type"`
	StackID          string         `json:"stack_id"`
	Source           RegisterSource `json:"source"`
}

// RegisterFinishResponse ...
type RegisterFinishResponse struct {
	Status                    string `json:"status"`
	BuildTriggerToken         string `json:"build_trigger_token"`
	BranchName                string `json:"branch_name"`
	IsWebhookAutoRegSupported bool   `json:"is_webhook_auto_reg_supported"`
}

// RegisterFinishURL ...
func RegisterFinishURL(appSlug string) string {
	return fmt.Sprintf(AppsServiceURL+"%s/finish", appSlug)
}

// RegisterFinish ...
func (s *AppService) RegisterFinish(params RegisterFinishParams) (*RegisterFinishResponse, error) {
	config, ok := configByProjectType[params.ProjectType]
	if !ok {
		return nil, fmt.Errorf("failed to select default config: unkown project type: %s", params.ProjectType)
	}

	type Params struct {
		RegisterFinishParams
		Config string         `json:"config"`
		Mode   string         `json:"mode"`
		Source RegisterSource `json:"source"`
	}
	p := Params{RegisterFinishParams: params}
	p.Mode = "manual"
	p.Config = config

	req, err := s.client.newRequest(http.MethodPost, RegisterFinishURL(s.Slug), p)
	if err != nil {
		return nil, err
	}
	var resp RegisterFinishResponse
	if err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
