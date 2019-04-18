package phases

import "fmt"

// IsPublic ...
func IsPublic() (bool, error) {
	fmt.Println("SetIsPublic")
	return false, nil
}
