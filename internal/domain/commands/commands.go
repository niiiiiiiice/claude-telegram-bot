package commands

type StartCommand struct {
	ChatID int64
	UserID int64
}

type HelpCommand struct {
	ChatID int64
	UserID int64
}

type StartBeginCommand struct {
	ChatID int64
	UserID int64
}

type EndChatCommand struct {
	ChatID int64
	UserID int64
}

type WhoAmICommand struct {
	ChatID    int64
	UserID    int64
	Username  string
	FirstName string
	LastName  string
}

type ProcessMessageCommand struct {
	ChatID   int64
	UserID   int64
	Message  string
	Username string
}
