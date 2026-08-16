package main

import (
	"errors"
	"log"
	"regexp"
)

type User struct {
	ID    int
	Name  string
	Email string
}

func NewUser(name, email string) (*User, error) {
	if !isValidEmail(email) {
		return nil, errors.New("Email have invalid pattern")
	}
	return &User{ID: 0, Name: name, Email: email}, nil
}

func (u *User) UpdateEmail(email string) error {
	if !isValidEmail(email) {
		return errors.New("Email have invalid pattern")
	}
	u.Email = email
	return nil
}

func GetUserByID(users []User, id int) (*User, error) {
	for _, user := range users {
		if user.ID == id {
			return &user, nil
		}
	}
	return nil, errors.New("User not found")
}

func isValidEmail(email string) bool {
	emailRegexp := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegexp.MatchString(email)
}

func main() {
	josh, _ := NewUser("Josh", "Josh@mail.eu")
	log.Printf("User is created: %+v\n", *josh)
	users := []User{*josh}
	_, error1 := GetUserByID(users, 1)
	log.Printf("Found error: %+v\n", error1)
	_, error2 := NewUser("David", "David.eu")
	log.Printf("Found error: %+v\n", error2)
}
