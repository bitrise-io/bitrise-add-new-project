package httputil

import (
	"net/http"
	"net/http/httputil"

	"github.com/bitrise-io/go-utils/log"
)

// PrintRequest ...
func PrintRequest(request *http.Request) error {
	if request == nil {
		return nil
	}

	dump, err := httputil.DumpRequest(request, true)
	if err != nil {
		return err
	}

	log.Debugf("%s", dump)

	return nil
}

// PrintResponse ...
func PrintResponse(response *http.Response) error {
	if response == nil {
		return nil
	}

	dump, err := httputil.DumpResponse(response, true)
	if err != nil {
		return err
	}
	log.Debugf("%s", dump)

	return nil
}
