package coap_test

import (
	"testing"

	coap "github.com/Kulak/go-coap"

	"time"

	"math"
)

func TestTimeoutDefaults(t *testing.T) {
	testTimeout(t, coap.DefaultResponseTimeout)
}

func TestTimeout6sec(t *testing.T) {
	testTimeout(t, 6*time.Second)
}

func testTimeout(t *testing.T, expectedTimeout time.Duration) {
	// send to IP address without CoAP EP
	req := coap.Message{
		Type:      coap.Confirmable,
		Code:      coap.GET,
		MessageID: 12345,
		Payload:   []byte("undeliverable"),
	}
	req.SetOption(coap.MaxAge, 3)
	req.SetPathString("/some/path")

	c, err := coap.DialWithTimeout("udp", "192.168.254.254:5683", expectedTimeout)
	if err != nil {
		t.Errorf("Error dialing: %v", err)
		return
	}

	start := time.Now()
	_, err = c.Send(req)
	end := time.Now()
	timeout := end.Sub(start)
	if err != nil {
		// that's not going to get printed unless test fails
		t.Logf("Error sending request: %v", err)
		if math.Abs(timeout.Seconds()-expectedTimeout.Seconds()) > 0.1 {
			t.Fatalf("Expected timeout %v, got %v", expectedTimeout, timeout)
		}
		return
	}
	t.Fatal("Response shall timeout.")
}
