package serviceLogs

import (
	"context"
	"fmt"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/projectService"
)

func (h *Handler) Run(ctx context.Context, config RunConfig) error {
	inputs, err := h.checkInputValues(config)
	if err != nil {
		return err
	}

	projectId, err := projectService.GetProjectId(ctx, h.apiGrpcClient, config.ProjectNameOrId, h.sdkConfig)
	if err != nil {
		return err
	}

	serviceName, source, containerIndex, err := h.getNameSourceContainerId(config)
	if err != nil {
		return err
	}

	service, err := projectService.GetServiceStack(ctx, h.apiGrpcClient, projectId, serviceName)
	if err != nil {
		return err
	}

	serviceTypeCategory := service.GetServiceStackTypeInfo().GetServiceStackTypeCategory().String()

	if serviceTypeCategory != USER {
		return fmt.Errorf("%s", i18n.LogRuntimeOnly)
	}
	serviceId := service.GetId()
	containerId := ""
	// defined by user, can be 1 or higher
	if containerIndex > 0 {
		containerId, err = h.getContainerId(ctx, h.sdkConfig, serviceId, containerIndex)
		if err != nil {
			return err
		}
	}

	logServiceId := serviceId
	if source == BUILD {
		logServiceId, err = h.getAppVersionServiceId(ctx, h.sdkConfig, serviceId)
		if err != nil {
			return err
		}
	}

	if err = h.printLogs(ctx, config, inputs, containerId, logServiceId, projectId); err != nil {
		return err
	}

	return nil
}
