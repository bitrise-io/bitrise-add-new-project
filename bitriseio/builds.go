package bitriseio

import (
	"fmt"
	"net/http"
)

// BuildParams ...
type BuildParams struct {
	Branch     string `json:"branch,omitempty"`
	WorkflowID string `json:"workflow_id,omitempty"`
}

// TriggerBuildParams ...
type TriggerBuildParams struct {
	BuildParams BuildParams `json:"build_params,omitempty"`
}

// TriggerBuildURL ...
func TriggerBuildURL(appSlug string) string {
	return fmt.Sprintf(AppsServiceURL+"%s/builds", appSlug)
}

// TriggerBuild ...
func (s *AppsService) TriggerBuild(appSlug string, params TriggerBuildParams) error {
	req, err := s.client.newRequest(http.MethodPost, TriggerBuildURL(appSlug), params)
	if err != nil {
		return err
	}
	return s.client.do(req, nil)
}
