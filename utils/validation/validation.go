// validation validates given messages for a given intent
package validation

import "strings"

// Validate an intent for confirmation
func ValidateConfirmation(message string) bool {
	for _, word := range yesWords {
		if strings.Contains(message, word) {
			return true
		}
	}

	return false
}

// Validate an intent for denial
func ValidateDenial(message string) bool {
	for _, word := range noWords {
		if strings.Contains(message, word) {
			return true
		}
	}

	return false
}

// Validate an intent to stop
func ValidateStop(message string) bool {
	for _, word := range stopWords {
		if strings.Contains(message, word) {
			return true
		}
	}

	return false
}
