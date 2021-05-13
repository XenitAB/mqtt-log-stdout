package message

import "fmt"

// Options takes the input configuration for the message client
type Options struct{}

type client struct{}

// Client interface
type Client interface {
	Print(m string)
}

// NewClient returns a Client interface
func NewClient(opts Options) Client {
	return &client{}
}

// Print takes a string and prints it to stdout
func (client *client) Print(m string) {
	fmt.Println(m)
}
