package main

// OffCommand is a concrete command
type OffCommand struct {
	device Device
}

// execute calls the off method on the device
func (c *OffCommand) execute() {
	c.device.off()
}
