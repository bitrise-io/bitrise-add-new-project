package phases

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type visibilityOption struct{
	Name string
	IsPublic bool
}

// IsPublic ...
func IsPublic() (bool, error) {
	options := []visibilityOption{
		visibilityOption{"Private", false},
		visibilityOption{"Public", true},
	}

	fmt.Println("SET THE PRIVACY OF THE APP")
	for i, opt := range options {
		fmt.Printf("%d) %s", i + 1, opt.Name)
		fmt.Println()
	}

	var choice int
	for !isValid(choice, len(options)) {
		fmt.Print("CHOOSE THE VISIBILITY: ")

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("error reading choice from stdin: %s", err)
			fmt.Println()
			continue
		}
		
		choice, err = strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			fmt.Printf("error reading choice from stdin: %s", err)
			fmt.Println()
			continue
		} else if !isValid(choice, len(options)) {
			fmt.Printf("invalid choice")
			fmt.Println()
			continue
		} else {
			break
		}
	}

	fmt.Printf("your choice was %s", options[choice - 1].Name)
	fmt.Println()

	return options[choice - 1].IsPublic, nil
}
