package phases

import "fmt"

func IsPublic() (bool, error) {
	fmt.Println("SetIsPublic")
	return false, nil
}
