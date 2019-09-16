package chantest

import "testing"

func TestCanUseT(t *testing.T) {
	var _ TestingT = t
}
