package main

// UserIterator is the iterator for the user collection
type UserIterator struct {
	index int
	users []*User
}

// hasNext checks if there are more users in the collection
func (u *UserIterator) hasNext() bool {
	return u.index < len(u.users)
}

// getNext returns the next user in the collection
func (u *UserIterator) getNext() *User {
	if u.hasNext() {
		user := u.users[u.index]
		u.index++
		return user
	}
	return nil
}
