package pool

import (
	"fmt"
	"net/netip"

	"go4.org/netipx"
)

func makeIPSet(prefixes []string) (*netipx.IPSet, error) {
	var b netipx.IPSetBuilder
	for _, v := range prefixes {
		p, err := netip.ParsePrefix(v)
		if err != nil {
			return nil, fmt.Errorf("failed to parse prefix %s, %w", v, err)
		}

		b.AddPrefix(p)
	}

	return b.IPSet()
}
