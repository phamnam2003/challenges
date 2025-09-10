package main

// OnCommand is a concrete command
type OnCommand struct {
	device Device
}

// execute calls the on method on the device
func (c *OnCommand) execute() {
	c.device.on()
}
