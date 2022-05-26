package cmd

import (
	"context"
	"time"

	"github.com/zerops-io/zcli/src/cliAction/startStopDelete"
	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/httpClient"

	"github.com/spf13/cobra"
)

func projectStopCmd() *cobra.Command {
	cmdStop := &cobra.Command{
		Use:          "stop [projectName]",
		Short:        i18n.CmdProjectStop,
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			regSignals(cancel)

			storage, err := createCliStorage()
			if err != nil {
				return err
			}

			region, err := createRegionRetriever(ctx)
			if err != nil {
				return err
			}

			reg, err := region.RetrieveFromFile()
			if err != nil {
				return err
			}

			apiClientFactory := business.New(business.Config{
				CaCertificateUrl: reg.CaCertificateUrl,
			})
			apiGrpcClient, closeFunc, err := apiClientFactory.CreateClient(
				ctx,
				reg.GrpcApiAddress,
				getToken(storage),
			)
			if err != nil {
				return err
			}
			defer closeFunc()

			client := httpClient.New(ctx, httpClient.Config{
				HttpTimeout: time.Minute * 15,
			})

			handler := startStopDelete.New(startStopDelete.Config{}, client, apiGrpcClient)

			cmdData := startStopDelete.CmdType{
				Start:   i18n.ProjectStop,
				Finish:  i18n.ProjectStopped,
				Execute: handler.ProjectStop,
			}

			return handler.Run(ctx, startStopDelete.RunConfig{
				ProjectName: args[0],
				ParentCmd:   constants.Project,
				Confirm:     true,
				CmdData:     cmdData,
			})
		},
	}
	return cmdStop
}
