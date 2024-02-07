package commands

// Command interface for all commands
type Command interface {
	Execute(args []string) error
	Usage() string
	Name() string
}

// Commands holds the registered commands
var Commands []Command
