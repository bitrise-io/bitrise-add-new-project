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
		t.Fatalf("Error getting current stack list from GitHub: %s", err)
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading stack info from GitHub response: %s", err)
	}

	var rb ResponseBody
	if err := json.Unmarshal(bytes, &rb); err != nil {
		t.Fatalf("Error unmarshalling stack data from string (%s): %s", bytes, err)
	}

	if len(config.Stacks) != len(rb) {
		t.Fatalf("Stack list changed")
	}

	for _, e := range rb {
		trimmed := strings.TrimSuffix(e.Name, ".log")
		if !sliceutil.IsStringInSlice(trimmed, config.Stacks) {
			t.Fatalf("Stack list changed")
		}
	}

}
