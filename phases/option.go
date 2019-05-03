package phases

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bitrise-io/go-utils/log"
)

type option struct {
	title        string
	valueOptions []string
	action       func(string) *option
}

func (o *option) run() {
	answer := ask(o.title, o.valueOptions...)
	if o.action != nil {
		if nextOption := o.action(answer); nextOption != nil {
			nextOption.run()
		}
	}
}

func ask(title string, options ...string) string {
	if len(options) == 1 {
		return options[0]
	}

	fmt.Print(strings.TrimSuffix(title, ":") + ":")

	if len(options) == 0 {
		fmt.Print(" ")
		for {
			input, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				log.Errorf("Error: failed to read input value")
				continue
			}
			fmt.Println()
			return strings.TrimSpace(input)
		}
	}

	fmt.Println()
	for i, option := range options {
		log.Printf("(%d) %s", i+1, option)
	}

	for {
		fmt.Print("Option number: ")
		answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Errorf("Error: failed to read input value")
			continue
		}
		optionNo, err := strconv.Atoi(strings.TrimSpace(answer))
		if err != nil {
			log.Errorf("Error: failed to parse option number, pick a number from 1-%d", len(options))
			continue
		}
		if optionNo-1 < 0 || optionNo-1 >= len(options) {
			log.Errorf("Error: invalid option number, pick a number 1-%d", len(options))
			continue
		}
		fmt.Println()
		return options[optionNo-1]
	}
}
