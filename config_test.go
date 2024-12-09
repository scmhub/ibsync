package ibsync

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := NewConfig()

	if config.Host != HOST {
		t.Errorf("expected default Host to be %s, got %s", HOST, config.Host)
	}

	if config.Port != PORT {
		t.Errorf("expected default Port to be %d, got %d", PORT, config.Port)
	}

	if config.ClientID < 1 || config.ClientID > 999999 {
		t.Errorf("expected default ClientID to be between 1 and 999999, got %d", config.ClientID)
	}

	if !config.InSync {
		t.Errorf("expected default InSync to be true, got %v", config.InSync)
	}

	if config.Timeout != TIMEOUT {
		t.Errorf("expected default Timeout to be %v, got %v", TIMEOUT, config.Timeout)
	}
}

func TestWithHost(t *testing.T) {
	config := NewConfig(WithHost("192.168.1.1"))

	if config.Host != "192.168.1.1" {
		t.Errorf("expected Host to be %s, got %s", "192.168.1.1", config.Host)
	}
}

func TestWithPort(t *testing.T) {
	config := NewConfig(WithPort(4001))

	if config.Port != 4001 {
		t.Errorf("expected Port to be %d, got %d", 4001, config.Port)
	}
}

func TestWithClientID(t *testing.T) {
	config := NewConfig(WithClientID(12345))

	if config.ClientID != 12345 {
		t.Errorf("expected ClientID to be %d, got %d", 12345, config.ClientID)
	}
}

func TestWithClientZero(t *testing.T) {
	config := NewConfig(WithClientZero())

	if config.ClientID != 0 {
		t.Errorf("expected ClientID to be 0, got %d", config.ClientID)
	}
}

func TestWithoutSync(t *testing.T) {
	config := NewConfig(WithoutSync())

	if config.InSync {
		t.Errorf("expected InSync to be false, got %v", config.InSync)
	}
}

func TestWithTimeout(t *testing.T) {
	customTimeout := 15 * time.Second
	config := NewConfig(WithTimeout(customTimeout))

	if config.Timeout != customTimeout {
		t.Errorf("expected Timeout to be %v, got %v", customTimeout, config.Timeout)
	}
}
