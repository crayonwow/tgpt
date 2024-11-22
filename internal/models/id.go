package models

type ID string

func (t ID) String() string {
	return string(t)
}

type UserID struct{ ID }
