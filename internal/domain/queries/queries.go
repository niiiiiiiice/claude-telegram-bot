package queries

type GetSessionQuery struct {
	ChatID int64
	UserID int64
}

type IsSessionActiveQuery struct {
	ChatID int64
	UserID int64
}
