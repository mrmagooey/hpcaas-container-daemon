package container

import "fmt"

var containerNamePrefix = "container"

func generateContainerName(id int) string {
	return fmt.Sprintf("%s_%d", containerNamePrefix, id)
}
