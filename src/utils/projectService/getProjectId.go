package projectService

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/sdkConfig"
)

func GetProjectId(ctx context.Context, apiGrpcClient business.ZeropsApiProtocolClient, projectNameOrId string, sdkConfig sdkConfig.Config) (string, error) {
	if projectNameOrId == "" {
		return "", errors.New(i18n.ProjectNameOrIdEmpty)
	}

	projects, err := getByName(ctx, apiGrpcClient, projectNameOrId)
	if err != nil {
		return "", err
	}

	if len(projects) > 1 {
		return "", getProjectSameNameErr(projects)
	}

	if len(projects) == 1 {
		return projects[0].GetId(), nil
	}

	projectId, err := getById(ctx, sdkConfig, projectNameOrId)
	if err != nil {
		return "", err
	}

	return projectId, nil
}

func getByName(ctx context.Context, apiGrpcClient business.ZeropsApiProtocolClient, projectName string) ([]*business.Project, error) {
	projectsResponse, err := apiGrpcClient.GetProjectsByName(ctx, &business.GetProjectsByNameRequest{
		Name: projectName,
	})
	if err := proto.BusinessError(projectsResponse, err); err != nil {
		return nil, err
	}
	projects := projectsResponse.GetOutput().GetProjects()
	return projects, nil
}

// return project IDs hint for projects with the same name
func getProjectSameNameErr(projects []*business.Project) error {
	out := make([]string, len(projects))
	for i, p := range projects {
		out[i] = p.GetId()
	}

	idList := strings.Join(out, ", ")
	errMsg := fmt.Errorf("%s\n%s%s", i18n.ProjectsWithSameName, i18n.AvailableProjectIds, idList)

	return errMsg
}
