package phases

import "fmt"

// Register ...
func Register(progress Progress) error {
	fmt.Println("Register")
	return nil
}
