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
func (s *AppService) BitriseYML(params BitriseYMLParams) error {
	req, err := s.client.newRequest(http.MethodPost, BitriseYMLURL(s.Slug), params)
	if err != nil {
		return err
	}
	return s.client.do(req, nil)
}
