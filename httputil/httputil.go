package httputil

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	"github.com/bitrise-io/go-utils/log"
)

// ReadBodyAndRestoreReadCloser ...
func ReadBodyAndRestoreReadCloser(body *io.ReadCloser) (string, error) {
	buf, err := ioutil.ReadAll(*body)
	*body = ioutil.NopCloser(bytes.NewBuffer(buf))

	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func printHeaders(header http.Header) {
	log.Debugf("Header:")
	for key, value := range header {
		log.Debugf("%s: %s", key, value)
	}
}

func printBody(response *http.Response) error {
	log.Debugf("Body:")

	bodyBytes, err := ReadBodyAndRestoreReadCloser(&response.Body)
	if err != nil {
		return err
	}

	log.Debugf("(%s)", bodyBytes)

	return nil
}

// PrintRequest ...
func PrintRequest(request *http.Request) error {
	if request == nil {
		return nil
	}

	requestBytes, err := httputil.DumpRequest(request, true)
	if err != nil {
		return err
	}

	log.Debugf("%s", requestBytes)

	return nil
}

// PrintResponse ...
func PrintResponse(response *http.Response) error {
	if response == nil {
		return nil
	}

	log.Debugf("StatusCode: %d", response.StatusCode)

	printHeaders(response.Header)

	err := printBody(response)
	if err != nil {
		return err
	}

	return nil
}
