package bitriseio

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/xcode-project/serialized"
)

// UploadKeystoreParams ...
type UploadKeystoreParams struct {
	Alias          string `json:"alias,omitempty"`
	Password       string `json:"password,omitempty"`
	KeyPassword    string `json:"private_key_password,omitempty"`
	UploadFileName string `json:"upload_file_name,omitempty"`
	UploadFileSize int64  `json:"upload_file_size,omitempty"`
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
func (s *AppsService) UploadKeystore(appSlug, pth string, params UploadKeystoreParams) error {
	f, err := os.Open(pth)
	if err != nil {
		return err
	}

	i, err := f.Stat()
	if err != nil {
		return err
	}

	name := filepath.Base(f.Name())
	name = strings.TrimSuffix(name, filepath.Ext(name))

	params.UploadFileSize = i.Size()
	params.UploadFileName = name

	// register keystore
	req, err := s.client.newRequest(http.MethodPost, UploadKeystoreURL(appSlug), params)
	if err != nil {
		return err
	}
	var resp serialized.Object
	if err := s.client.do(req, &resp); err != nil {
		return err
	}

	data, err := resp.Object("data")
	if err != nil {
		return err
	}

	uploadURL, err := data.String("upload_url")
	if err != nil {
		return err
	}

	uploadSlug, err := data.String("slug")
	if err != nil {
		return err
	}

	content, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	// upload keystore
	req, err = http.NewRequest(http.MethodPut, uploadURL, bytes.NewReader(content))
	if err != nil {
		return err
	}

	if err := s.client.do(req, nil); err != nil {
		return err
	}

	// confirm upload
	req, err = s.client.newRequest(http.MethodPost, UploadKeystoreConfirmURL(appSlug, uploadSlug), nil)
	if err != nil {
		return err
	}

	return s.client.do(req, nil)
}
