package phases

import "fmt"

func Register(progress Progress) error {
	fmt.Println("Register")
	return nil
}
