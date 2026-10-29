package automation

import (
	"context"
	"fmt"
)

// ValidateCandidate checks source provenance without requiring automation credentials.
func ValidateCandidate(ctx context.Context, r Runner, dir, base string) error {
	if base != "refs/remotes/origin/main" {
		return fmt.Errorf("validation base must be freshly fetched origin/main")
	}
	c := commands{ctx, r, dir}
	p, e := c.pinAt("HEAD")
	if e != nil {
		return e
	}
	old, e := c.pinAt(base)
	if e != nil {
		return e
	}
	if old != nil {
		if e = c.verifyPin(*old, base); e != nil {
			return e
		}
	}
	if p == nil {
		if old != nil {
			return fmt.Errorf("candidate removes upstream source pin")
		}
		return nil
	}
	if e = c.verifyPin(*p, "HEAD"); e != nil {
		return e
	}
	return checkAdvance(old, *p)
}
