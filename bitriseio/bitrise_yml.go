package bitriseio

import (
	"fmt"
	"net/http"
)

// BitriseYMLParams ...
type BitriseYMLParams struct {
	AppConfigDatastoreYAML string `json:"app_config_datastore_yaml,omitempty"`
}

// BitriseYMLURL ...
func BitriseYMLURL(appSlug string) string {
	return fmt.Sprintf(AppsServiceURL+"%s/bitrise.yml", appSlug)
}

// BitriseYML ...
func (s *AppsService) BitriseYML(appSlug string, params BitriseYMLParams) error {
	req, err := s.client.newRequest(http.MethodPost, BitriseYMLURL(appSlug), params)
	if err != nil {
		return err
	}
	return s.client.do(req, nil)
}
