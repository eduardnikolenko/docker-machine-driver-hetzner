package main

import "github.com/docker/machine/libmachine/drivers/plugin"

var (
	// Version ...
	Version string
)

func main() {
	plugin.RegisterDriver(NewDriver())
}
