package projectService

import (
	"context"
	"errors"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
)

func GetServiceStack(ctx context.Context, apiGrpcClient business.ZeropsApiProtocolClient, projectId string, serviceName string) (*business.GetServiceStackByNameResponseDto, error) {

	if serviceName == "" {
		return nil, errors.New(i18n.ServiceNameIsEmpty)
	}

	response, err := apiGrpcClient.GetServiceStackByName(ctx, &business.GetServiceStackByNameRequest{
		ProjectId: projectId,
		Name:      serviceName,
	})
	if err := proto.BusinessError(response, err); err != nil {
		return nil, err
	}

	return response.GetOutput(), nil
}

func GetServiceId(ctx context.Context, apiGrpcClient business.ZeropsApiProtocolClient, projectId string, serviceName string) (string, error) {
	service, err := GetServiceStack(ctx, apiGrpcClient, projectId, serviceName)
	if err != nil {
		return "", err
	}

	id := service.GetId()

	if len(id) == 0 {
		return "", errors.New(i18n.ServiceNotFound)
	}

	return id, nil
}
