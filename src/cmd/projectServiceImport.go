package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/yamlReader"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/types"
)

const serviceImportArgName = "importYamlPath"

func projectServiceImportCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("service-import").
		Short(i18n.T(i18n.CmdServiceImport)).
		ScopeLevel(cmdBuilder.Project).
		Arg(serviceImportArgName).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			uxBlocks := cmdData.UxBlocks

			yamlContent, err := yamlReader.ReadContent(uxBlocks, cmdData.Args[serviceImportArgName][0], "./")
			if err != nil {
				return err
			}

			importServiceResponse, err := cmdData.RestApiClient.PostServiceStackImport(
				ctx,
				body.ServiceStackImport{
					ProjectId: cmdData.Project.ID,
					Yaml:      types.Text(yamlContent),
				},
			)
			if err != nil {
				return err
			}

			responseOutput, err := importServiceResponse.Output()
			if err != nil {
				return err
			}

			var processes []uxHelpers.Process
			for _, service := range responseOutput.ServiceStacks {
				for _, process := range service.Processes {
					processes = append(processes, uxHelpers.Process{
						Id:                  process.Id,
						RunningMessage:      service.Name.String() + ": " + process.ActionName.String(),
						ErrorMessageMessage: service.Name.String() + ": " + process.ActionName.String(),
						SuccessMessage:      service.Name.String() + ": " + process.ActionName.String(),
					})
				}
			}

			uxBlocks.PrintLine(i18n.T(i18n.ServiceCount, len(responseOutput.ServiceStacks)))
			uxBlocks.PrintLine(i18n.T(i18n.QueuedProcesses, len(processes)))

			err = uxHelpers.ProcessCheckWithSpinner(ctx, cmdData.UxBlocks, cmdData.RestApiClient, processes)
			if err != nil {
				return err
			}

			uxBlocks.PrintInfoLine(i18n.T(i18n.ServiceImported))

			return nil
		})
}