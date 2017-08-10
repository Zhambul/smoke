package domain

type Account struct {
	Id         int
	FirstName  string
	LastName   string
	ChatId     int
}

type Group struct {
	Id             int
	Name           string
	CreatorAccount *Account
	Accounts       []*Account
	UUID           string
}

type Message struct {
	Id         int
	Text       string
	From       Account
	Recipients []*Account
}