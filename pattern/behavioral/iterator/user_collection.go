package main

// UserCollection is the struct that holds user information
type UserCollection struct {
	users []*User
}

// createIterator creates an iterator for the user collection
func (u *UserCollection) createIterator() Iterator {
	return &UserIterator{
		users: u.users,
	}
}
