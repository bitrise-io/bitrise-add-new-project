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
	fmt.Print("CHOOSE THE VISIBILITY: ")

	// scan for input
	reader := bufio.NewReader(os.Stdin)

	input, err := reader.ReadString('\n')
	if err != nil {
		// todo
	}
	
	choice, err := strconv.Atoi(strings.Trim(input, "\n"))
	if err != nil {
		// todo
	}

	fmt.Printf("your choice was %s", options[choice - 1].Name)
	fmt.Println()

	return options[choice - 1].IsPublic, nil
}
