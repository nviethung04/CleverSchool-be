package command

type Command interface {
	Execute() bool
}

type CreateUserCommand struct {
	Title string
}

func (c CreateUserCommand) Execute() bool {
	return true
}

type DeleteUserCommand struct {
	ID int
}

func (d DeleteUserCommand) Execute() bool {
	return true
}

// ===== Reindex Questions Command =====
// go run cmd/reindex_questions.go (or wire via routes/console if needed)
