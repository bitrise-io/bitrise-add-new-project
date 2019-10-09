package maintenance

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise-add-new-project/config"
	"github.com/bitrise-io/go-utils/sliceutil"
)

type ResponseBody []DirectoryEntry
type DirectoryEntry struct {
	Name string `json:"name"`
}

func TestStackChange(t *testing.T) {
	resp, err := http.Get("https://api.github.com/repos/bitrise-io/bitrise.io/contents/system_reports")
	if err != nil {

	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {

	}

	var rb ResponseBody
	if err := json.Unmarshal(bytes, &rb); err != nil {

	}

	changed := false
	for _, e := range rb {
		trimmed := strings.TrimSuffix(e.Name, ".log")
		if !sliceutil.IsStringInSlice(trimmed, config.Stacks) {
			changed = true
		}
	}

	if changed {
		t.Logf("Stack list changed.")
	}

}
