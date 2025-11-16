package domain

type User struct {
	Id       string
	Name     string
	Team     string
	IsActive bool
}

type Team struct {
	Name    string
	Members []User
}
