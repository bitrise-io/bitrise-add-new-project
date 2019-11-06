package maintenance

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise-add-new-project/config"
)

type report struct {
	Name string `json:"name"`
}

type systemReports []report

func (reports systemReports) Stacks() (s []string) {
	for _, report := range reports {
		s = append(s, strings.TrimSuffix(report.Name, ".log"))
	}
	return
}

func TestStackChange(t *testing.T) {
	resp, err := http.Get("https://api.github.com/repos/bitrise-io/bitrise.io/contents/system_reports?access_token=" + os.Getenv("GIT_BOT_USER_ACCESS_TOKEN"))
	if err != nil {
		t.Fatalf("Error getting current stack list from GitHub: %s", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("Error closing response body")
		}
	}()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading stack info from GitHub response: %s", err)
	}

	var reports systemReports
	if err := json.Unmarshal(bytes, &reports); err != nil {
		t.Fatalf("Error unmarshalling stack data from string (%s): %s", bytes, err)
	}

	if expected := reports.Stacks(); !reflect.DeepEqual(expected, config.Stacks()) {
		t.Fatalf("Stack list changed, current: %v, expecting: %v", config.Stacks(), expected)
	}
}
