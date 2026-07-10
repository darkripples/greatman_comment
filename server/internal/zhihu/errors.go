package zhihu

import "strings"

const APIErrorCodeRateLimit = 30001

func IsRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "code 30001") || strings.Contains(msg, "limit exceeded")
}

func IsDayLimitExceeded(err error) bool {
	if !IsRateLimitError(err) {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "day limit")
}

func IsSecondLimitExceeded(err error) bool {
	if !IsRateLimitError(err) {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "second limit")
}
