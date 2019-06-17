package bitriseio

import (
	"fmt"
	"net/http"
)

// TriggerBuildURL ...
func TriggerBuildURL(appSlug string) string {
	return fmt.Sprintf(AppsServiceURL+"%s/builds", appSlug)
}

// TriggerBuild ...
func (s *AppService) TriggerBuild(workflowID, branch string) error {
	type HookInfo struct {
		Type string `json:"type"`
	}
	type BuildParams struct {
		WorkflowID string `json:"workflow_id"`
		Branch     string `json:"branch"`
	}
	type Params struct {
		BuildParams BuildParams `json:"build_params"`
		HookInfo    HookInfo    `json:"hook_info"`
	}
	p := Params{
		BuildParams: BuildParams{
			WorkflowID: workflowID,
			Branch:     branch,
		},
		HookInfo: HookInfo{
			Type: "bitrise",
		},
	}
	req, err := s.client.newRequest(http.MethodPost, TriggerBuildURL(s.Slug), p)
	if err != nil {
		return err
	}
	return s.client.do(req, nil)
}
