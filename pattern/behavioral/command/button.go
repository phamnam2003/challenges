package main

// Button is the invoker
type Button struct {
	command Command
}

// press invokes the command's execute method
func (b *Button) press() {
	b.command.execute()
}
