package cmd

import (
	"context"
	"time"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/proto/business"

	"github.com/zerops-io/zcli/src/cliAction/startStopDelete"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/httpClient"

	"github.com/spf13/cobra"
)

func serviceStartCmd() *cobra.Command {
	cmdStart := &cobra.Command{
		Use:          "start [projectName] [serviceName]",
		Short:        i18n.CmdServiceStart,
		Args:         cobra.MinimumNArgs(2),
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
				Start:   i18n.ServiceStart,
				Finish:  i18n.ServiceStarted,
				Execute: handler.ServiceStart,
			}

			return handler.Run(ctx, startStopDelete.RunConfig{
				ProjectName: args[0],
				ServiceName: args[1],
				ParentCmd:   constants.Service,
				Confirm:     true,
				CmdData:     cmdData,
			})
		},
	}
	return cmdStart
}
