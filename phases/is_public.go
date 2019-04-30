package phases

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bitrise-io/go-utils/log"
)

const (
	optPrivate = "Private"
	optPublic  = "Public"
)

// IsPublic returns the whether the Bitrise project
// should be public or not.
func IsPublic() (bool, error) {
	options := []string{optPrivate, optPublic}

	log.Infof("SET THE PRIVACY OF THE APP")
	log.Infof("==========================")
	for i, opt := range options {
		log.Printf("%d) %s", i+1, opt)
	}

	for {
		log.Warnf("CHOOSE THE VISIBILITY: ")

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Errorf("error reading choice from stdin: %s", err)
			continue
		}

		choice, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			log.Errorf("error reading choice from stdin: %s", err)
			continue
		} else if !isValid(choice, len(options)) {
			log.Errorf("invalid choice")
			continue
		}

		return options[choice-1] == optPublic, nil
	}

	return false, fmt.Errorf("invalid execution branch: unknown error")
}
