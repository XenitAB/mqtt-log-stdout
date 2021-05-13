package status

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Options takes the input configuration for the status client
type Options struct {
	ClientID string
}
type statusMessage struct {
	Timestamp    time.Time `json:"timestamp"`
	ClientID     string    `json:"client_id"`
	Message      string    `json:"message"`
	ErrorMessage string    `json:"error,omitempty"`
}

type client struct {
	clientID string
}

// Client interface
type Client interface {
	Print(m string, e error)
}

// NewClient returns a Client interface
func NewClient(opts Options) Client {
	return &client{
		clientID: opts.ClientID,
	}
}

// Print prints messages to stdout or stderr
func (s *client) Print(m string, e error) {
	output := os.Stdout
	errMsg := ""
	if e != nil {
		errMsg = e.Error()
		output = os.Stderr
	}

	status := statusMessage{
		Timestamp:    time.Now(),
		ClientID:     s.clientID,
		Message:      m,
		ErrorMessage: errMsg,
	}

	res, err := json.Marshal(status)
	if err != nil {
		res = []byte(fmt.Sprintf("{\"timestamp\":\"%s\",\"client_id\":\"%s\",\"message\":\"\",\"error\":\"json.Marshal(status) failed: %s\"}", status.Timestamp, status.ClientID, err.Error()))
		output = os.Stderr
	}

	fmt.Fprintf(output, "%s\n", string(res))
}
