package maintenance

import (
	"strings"
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
