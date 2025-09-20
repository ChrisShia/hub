package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	ErrDuplicateEmail = errors.New("duplicate email")
)

type Models struct {
	Users interface {
		Insert(user *User) error
		Update(user *User) error
		GetByEmail(email string) (*User, error)
		Delete(email string) error
	}
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users: &UserModel{db},
	}
}

func NewMockModels() *Models {
	return &Models{
		Users: MockUserModel{},
	}
}
