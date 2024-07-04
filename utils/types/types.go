// Package types is used to hold global types used all throughout Horus
package types

/* ---- MODULE INTERFACE ---- */

// Module interface that the library uses to collect and use modules
type Module interface {
	Handler(function string, input *Input) any // Handle function calls to the module
}
