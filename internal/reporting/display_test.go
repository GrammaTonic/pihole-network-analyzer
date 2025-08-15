package reporting

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"pihole-analyzer/internal/types"
)

// captureOutput captures stdout output for assertions in tests
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = old
	return buf.String()
}

func TestDisplayResultsWithConfig_NoData(t *testing.T) {
	clientStats := map[string]*types.ClientStats{}
	// Should not panic or error
	DisplayResultsWithConfig(clientStats, nil)
}

func TestDisplayResultsWithConfig_SingleClient(t *testing.T) {
	clientStats := map[string]*types.ClientStats{
		"192.168.1.2": {
			IP:           "192.168.1.2",
			Hostname:     "test-host",
			TotalQueries: 42,
			IsOnline:     true,
		},
	}
	DisplayResultsWithConfig(clientStats, nil)
	// Check that the output contains the expected client IP
	output := captureOutput(func() {
		DisplayResultsWithConfig(clientStats, nil)
	})
	if !strings.Contains(output, "IP: 192.168.1.2") {
		t.Errorf("Expected output to contain 'IP: 192.168.1.2', got: %s", output)
	}
}

func TestDisplayResultsWithConfig_MultipleClients(t *testing.T) {
	clientStats := map[string]*types.ClientStats{
		"192.168.1.2": {
			IP:           "192.168.1.2",
			Hostname:     "host1",
			TotalQueries: 100,
			IsOnline:     true,
		},
		"192.168.1.3": {
			IP:           "192.168.1.3",
			Hostname:     "host2",
			TotalQueries: 50,
			IsOnline:     false,
		},
	}
	DisplayResultsWithConfig(clientStats, nil)
	// Check that both IPs are displayed in the correct context
	output := captureOutput(func() {
		DisplayResultsWithConfig(clientStats, nil)
	})
	if !strings.Contains(output, "IP: 192.168.1.2") {
		t.Errorf("Expected output to contain 'IP: 192.168.1.2', got: %s", output)
	}
	if !strings.Contains(output, "IP: 192.168.1.3") {
		t.Errorf("Expected output to contain 'IP: 192.168.1.3', got: %s", output)
	}
}
