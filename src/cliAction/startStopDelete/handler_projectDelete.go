package startStopDelete

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/processChecker"
)

func (h *Handler) ProjectDelete(ctx context.Context, projectId string, config RunConfig) error {

	if !config.Confirm {
		// run confirm dialogue
		shouldDelete := askForConfirmation(constants.Project)
		if !shouldDelete {
			fmt.Println(i18n.DelProjectCanceledByUser)
			return nil
		}
	}

	fmt.Println(i18n.DeleteProjectProcessInit)

	deleteProjectResponse, err := h.apiGrpcClient.DeleteProject(ctx, &business.DeleteProjectRequest{
		Id: projectId,
	})
	if err := proto.BusinessError(deleteProjectResponse, err); err != nil {
		return err
	}

	processId := deleteProjectResponse.GetOutput().GetId()

	err = processChecker.CheckProcess(ctx, processId, h.apiGrpcClient)
	if err != nil {
		return err
	}

	fmt.Println(constants.Success + i18n.DeleteProjectSuccess)

	return nil
}
