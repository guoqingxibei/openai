package util

import "testing"

func TestCountImageTokens(t *testing.T) {
	if CountImageTokens(150, 150) != 255 {
		t.Error()
	}

	if CountImageTokens(512, 512) != 255 {
		t.Error()
	}

	if CountImageTokens(5120, 1000) != 765 {
		t.Error()
	}

	if CountImageTokens(1000, 2000) != 1105 {
		t.Error()
	}

	if CountImageTokens(1000, 1000) != 765 {
		t.Error()
	}

	if CountImageTokens(2048, 2048) != 765 {
		t.Error()
	}

	// max value
	if CountImageTokens(2048, 768) != 1445 {
		t.Error()
	}
}
