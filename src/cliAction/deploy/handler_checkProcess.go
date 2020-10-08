package deploy

import (
	"context"
	"errors"
	"github.com/zerops-io/zcli/src/i18n"
	"time"

	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
)

func (h *Handler) checkProcess(ctx context.Context, processId string) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			getProcessResponse, err := h.apiGrpcClient.GetProcess(ctx, &zeropsApiProtocol.GetProcessRequest{
				Id: processId,
			})
			if err := utils.HandleGrpcApiError(getProcessResponse, err); err != nil {
				return err
			}

			processStatus := getProcessResponse.GetOutput().GetStatus()

			if processStatus == zeropsApiProtocol.ProcessStatus_PROCESS_STATUS_FINISHED {
				return nil
			}

			if !(processStatus == zeropsApiProtocol.ProcessStatus_PROCESS_STATUS_RUNNING ||
				processStatus == zeropsApiProtocol.ProcessStatus_PROCESS_STATUS_PENDING) {
				return errors.New(i18n.ProcessInvalidState + ": " + zeropsApiProtocol.ProcessStatus_name[int32(processStatus)])
			}
			time.Sleep(time.Second)
		}
	}
}
