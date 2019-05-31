package bitriseio

import (
	"fmt"
	"net/http"
)

// BitriseYMLURL ...
func BitriseYMLURL(appSlug string) string {
	return fmt.Sprintf(AppsServiceURL+"%s/bitrise.yml", appSlug)
}

// UploadBitriseYML ...
func (s *AppService) UploadBitriseYML(config string) error {
	type BitriseYMLParams struct {
		AppConfigDatastoreYAML string `json:"app_config_datastore_yaml"`
	}

	p := BitriseYMLParams{
		AppConfigDatastoreYAML: config,
	}

	req, err := s.client.newRequest(http.MethodPost, BitriseYMLURL(s.Slug), p)
	if err != nil {
		return err
	}
	return s.client.do(req, nil)
}
