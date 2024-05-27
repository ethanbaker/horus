// Package validation validates given messages for a given intent
package validation

// TODO: perform stronger word analysis than checking against word bank

import "strings"

// Validate an intent for confirmation
func ValidateConfirmation(message string) bool {
	for _, word := range yesWords {
		for _, w := range strings.Split(message, " ") {
			if w == word {
				return true
			}
		}
	}

	return false
}

// Validate an intent for denial
func ValidateDenial(message string) bool {
	for _, word := range noWords {
		for _, w := range strings.Split(message, " ") {
			if w == word {
				return true
			}
		}
	}

	return false
}

// Validate an intent to stop
func ValidateStop(message string) bool {
	for _, word := range stopWords {
		for _, w := range strings.Split(message, " ") {
			if w == word {
				return true
			}
		}
	}

	return false
}
