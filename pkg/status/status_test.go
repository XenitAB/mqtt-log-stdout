package status

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestPrint(t *testing.T) {
	statusClient := NewClient(Options{ClientID: "fake"})

	stdout := os.Stdout
	defer func() {
		os.Stdout = stdout
	}()

	r, w, err := os.Pipe()
	if err != nil {
		t.Errorf("Expected err to be nil: %q", err)
	}

	os.Stdout = w
	statusClient.Print("fake message", nil)
	w.Close()

	outputBytes, err := ioutil.ReadAll(r)
	if err != nil {
		t.Errorf("Expected err to be nil: %q", err)
	}

	output := string(outputBytes)

	if !strings.Contains(output, "\"timestamp\":") {
		t.Errorf("Expected output to contain '\"timestamp\":' but was: %q", output)
	}

	if !strings.Contains(output, "\"client_id\":\"fake\"") {
		t.Errorf("Expected output to contain '\"client_id\":\"fake\"' but was: %q", output)
	}

	if !strings.Contains(output, "\"message\":\"fake message\"") {
		t.Errorf("Expected output to contain '\"message\":\"fake message\"' but was: %q", output)
	}
}

func TestPrintErr(t *testing.T) {
	statusClient := NewClient(Options{ClientID: "fake"})

	stderr := os.Stderr
	defer func() {
		os.Stdout = stderr
	}()

	r, w, err := os.Pipe()
	if err != nil {
		t.Errorf("Expected err to be nil: %q", err)
	}

	os.Stderr = w
	statusClient.Print("fake message", fmt.Errorf("fake error"))
	w.Close()

	outputBytes, err := ioutil.ReadAll(r)
	if err != nil {
		t.Errorf("Expected err to be nil: %q", err)
	}

	output := string(outputBytes)

	if !strings.Contains(output, "\"timestamp\":") {
		t.Errorf("Expected output to contain '\"timestamp\":' but was: %q", output)
	}

	if !strings.Contains(output, "\"client_id\":\"fake\"") {
		t.Errorf("Expected output to contain '\"client_id\":\"fake\"' but was: %q", output)
	}

	if !strings.Contains(output, "\"message\":\"fake message\"") {
		t.Errorf("Expected output to contain '\"message\":\"fake message\"' but was: %q", output)
	}

	if !strings.Contains(output, "\"error\":\"fake error\"") {
		t.Errorf("Expected output to contain '\"error\":\"fake error\"' but was: %q", output)
	}
}
