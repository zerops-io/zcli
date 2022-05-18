package importProjectService

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
)

func (h *Handler) getClientId(ctx context.Context, config RunConfig) (string, error) {

	if len(config.ClientId) > 0 {
		return config.ClientId, nil
	}

	res, err := h.apiGrpcClient.GetUserInfo(ctx, &business.GetUserInfoRequest{})
	if err := proto.BusinessError(res, err); err != nil {
		return "", err
	}
	clients := res.GetOutput().GetClientUserList()

	if len(clients) == 0 {
		return "", errors.New(i18n.MissingClientId)
	}

	if len(clients) > 1 {
		var out []string

		for _, client := range clients {
			out = append(out, client.ClientId)
		}
		idList := strings.Join(out, ",")
		errMsg := fmt.Errorf("%s\n%s%s", i18n.MultipleClientIds, i18n.AvailableClientIds, idList)

		return "", errMsg
	}

	return clients[0].ClientId, nil
}
