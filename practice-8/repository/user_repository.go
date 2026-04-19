package repository

type User struct {
	ID    int
	Name  string
	Email string
}

type UserRepository interface {
	GetUserByID(id int) (*User, error)
	CreateUser(user *User) error
	// Task 2: new methods
	GetByEmail(email string) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(id int) error
}
