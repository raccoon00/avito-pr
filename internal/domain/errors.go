package domain

import "fmt"

type TeamExistsError struct {
	TeamName string
}

func (e *TeamExistsError) Error() string {
	return fmt.Sprintf("The team with a name %s already exists", e.TeamName)
}
