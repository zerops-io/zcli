package vpn

import (
	"context"
	"time"

	"github.com/zeropsio/zcli/src/nettools"
)

func (h *Handler) dnsIsAlive() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	err := nettools.Ping(ctx, "node1.master.core.zerops")
	if err != nil {
		return false, nil
	}
	return true, nil
}
