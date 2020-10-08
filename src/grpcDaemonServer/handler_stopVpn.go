package grpcDaemonServer

import (
	"context"

	"github.com/zerops-io/zcli/src/zeropsDaemonProtocol"
)

func (h *Handler) StopVpn(ctx context.Context, request *zeropsDaemonProtocol.StopVpnRequest) (*zeropsDaemonProtocol.StopVpnResponse, error) {

	err := h.vpn.StopVpn()
	if err != nil {
		return &zeropsDaemonProtocol.StopVpnResponse{}, err
	}

	return &zeropsDaemonProtocol.StopVpnResponse{}, nil
}
