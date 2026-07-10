package zhihu

import (
	"fmt"
	"testing"
)

func TestIsDayLimitExceeded(t *testing.T) {
	err := fmt.Errorf("zhihu api error (code %d): day limit exceeded", APIErrorCodeRateLimit)
	if !IsDayLimitExceeded(err) {
		t.Fatal("expected day limit")
	}
	if IsSecondLimitExceeded(err) {
		t.Fatal("unexpected second limit")
	}
}

func TestIsSecondLimitExceeded(t *testing.T) {
	err := fmt.Errorf("zhihu api error (code %d): second limit exceeded", APIErrorCodeRateLimit)
	if !IsSecondLimitExceeded(err) {
		t.Fatal("expected second limit")
	}
	if IsDayLimitExceeded(err) {
		t.Fatal("unexpected day limit")
	}
}
