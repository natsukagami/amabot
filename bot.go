// Package amabot provides a basic functioning AMA bot with question posting and queuing.
// It runs on one single server only.
package amabot

import "os"

// The owner ID.
// It gets initialized as an environment variable.
// Some init commands can only be sent from an owner.
var owner string

func init() {
	owner = os.Getenv("OWNER")
	if owner == "" {
		// No owner was found
		panic("Missing OWNER environment variable.")
	}
}
