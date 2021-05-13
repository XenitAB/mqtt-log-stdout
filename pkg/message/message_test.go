package message

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestPrint(t *testing.T) {
	messageClient := NewClient(Options{})

	stdout := os.Stdout
	defer func() {
		os.Stdout = stdout
	}()

	r, w, err := os.Pipe()
	if err != nil {
		t.Errorf("Expected err to be nil: %q", err)
	}

	os.Stdout = w
	messageClient.Print("fake message")
	w.Close()

	outputBytes, err := ioutil.ReadAll(r)
	if err != nil {
		t.Errorf("Expected err to be nil: %q", err)
	}

	output := string(outputBytes)

	if output != "fake message\n" {
		t.Errorf("Expected output to be '\"fake message\":' but was: %q", output)
	}
}
