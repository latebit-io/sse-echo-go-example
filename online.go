package main

type Online struct {
	users map[string]int
}

func NewOnline() *Online {
	return &Online{
		users: make(map[string]int),
	}
}

func (o *Online) AddUser(user string) {
	o.users[user]++
}

func (o *Online) RemoveUser(user string) {
	delete(o.users, user)
}

func (o *Online) GetUsers() []string {
	var users []string
	for user := range o.users {
		users = append(users, user)
	}
	return users
}
