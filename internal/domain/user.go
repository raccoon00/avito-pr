package domain

type User struct {
	Id       int
	Name     string
	Team     string
	IsActive bool
}

type Team struct {
	Name    string
	Members []User
}
