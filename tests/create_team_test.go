package tests

import "testing"

type Team struct {
	Name    string       `json:"team_name"`
	Members []TeamMember `json:"members"`
}

type TeamMember struct {
	Id   string `json:"user_id"`
	Name string `json:"user_name"`
}

func TestTeamAdd(t *testing.T) {

}
