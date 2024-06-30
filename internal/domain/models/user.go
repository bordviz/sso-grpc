package models

type User struct {
	ID       int64
	Name     string
	Email    string
	PassHash string
}

type UserRead struct {
	ID    int64
	Name  string
	Email string
}
