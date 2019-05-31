package bitriseio

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/log"
)

// UploadKeystoreParams ...
type UploadKeystoreParams struct {
	Alias       string `json:"alias"`
	Password    string `json:"password"`
	KeyPassword string `json:"private_key_password"`
}

// UploadKeystoreURL ...
func UploadKeystoreURL(appSlug string) string {
	return fmt.Sprintf(AppsServiceURL+"%s/android-keystore-files", appSlug)
}

// UploadKeystoreConfirmURL ...
func UploadKeystoreConfirmURL(appSlug, uploadSlug string) string {
	return fmt.Sprintf("%s/%s/uploaded", UploadKeystoreURL(appSlug), uploadSlug)
}

// UploadKeystore ...
func (s *AppService) UploadKeystore(pth string, params UploadKeystoreParams) error {
	f, err := os.Open(pth)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Debugf("failed to close keystore file: %s", err)
		}
	}()

	i, err := f.Stat()
	if err != nil {
		return err
	}

	name := filepath.Base(f.Name())
	name = strings.TrimSuffix(name, filepath.Ext(name))

	type Params struct {
		UploadKeystoreParams
		UploadFileName string `json:"upload_file_name"`
		UploadFileSize int64  `json:"upload_file_size"`
	}
	p := Params{UploadKeystoreParams: params}
	p.UploadFileSize = i.Size()
	p.UploadFileName = name

	// register keystore
	req, err := s.client.newRequest(http.MethodPost, UploadKeystoreURL(s.Slug), p)
	if err != nil {
		return err
	}

	type UploadKeystoreResponse struct {
		Data struct {
			UploadURL string `json:"upload_url"`
			Slug      string `json:"slug"`
		} `json:"data"`
	}

	var r UploadKeystoreResponse

	if err := s.client.do(req, &r); err != nil {
		return err
	}

	content, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	// upload keystore
	req, err = http.NewRequest(http.MethodPut, r.Data.UploadURL, bytes.NewReader(content))
	if err != nil {
		return err
	}

	if err := s.client.do(req, nil); err != nil {
		return err
	}

	// confirm upload
	req, err = s.client.newRequest(http.MethodPost, UploadKeystoreConfirmURL(s.Slug, r.Data.Slug), nil)
	if err != nil {
		return err
	}

	return s.client.do(req, nil)
}
