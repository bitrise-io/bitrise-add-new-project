package phases

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/bitrise-io/go-utils/log"
)

type visibilityOption struct {
	Name     string
	IsPublic bool
}

// IsPublic returns the whether the Bitrise project
// should be public or not.
func IsPublic() (bool, error) {
	options := []visibilityOption{
		visibilityOption{"Private", false},
		visibilityOption{"Public", true},
	}

	log.Infof("SET THE PRIVACY OF THE APP")
	log.Infof("==========================")
	for i, opt := range options {
		log.Printf("%d) %s", i+1, opt.Name)
	}

	var choice int
	for !isValid(choice, len(options)) {
		log.Warnf("CHOOSE THE VISIBILITY: ")

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Warnf("error reading choice from stdin: %s", err)
			continue
		}

		choice, err = strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			log.Warnf("error reading choice from stdin: %s", err)
			continue
		} else if !isValid(choice, len(options)) {
			log.Warnf("invalid choice")
			continue
		} else {
			break
		}
	}

	return options[choice-1].IsPublic, nil
}
