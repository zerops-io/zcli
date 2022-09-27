package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/zeropsio/zcli/src/cliAction/importProjectService"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zeropsio/zcli/src/utils/httpClient"
	"github.com/zeropsio/zcli/src/utils/sdkConfig"
)

func serviceImportCmd() *cobra.Command {
	cmdImport := &cobra.Command{
		Use:          "import projectNameOrId pathToImportFile [flags]",
		Short:        i18n.CmdServiceImport,
		Args:         ExactNArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			regSignals(cancel)

			storage, err := createCliStorage()
			if err != nil {
				return err
			}
			token, err := getToken(storage)
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

			apiClientFactory := zBusinessZeropsApiProtocol.New(zBusinessZeropsApiProtocol.Config{
				CaCertificateUrl: reg.CaCertificateUrl,
			})
			apiGrpcClient, closeFunc, err := apiClientFactory.CreateClient(
				ctx,
				reg.GrpcApiAddress,
				token,
			)
			if err != nil {
				return err
			}
			defer closeFunc()

			client := httpClient.New(ctx, httpClient.Config{
				HttpTimeout: time.Minute * 15,
			})

			return importProjectService.New(
				importProjectService.Config{}, client, apiGrpcClient, sdkConfig.Config{Token: token, RegionUrl: reg.RestApiAddress},
			).Import(ctx, importProjectService.RunConfig{
				WorkingDir:      constants.WorkingDir,
				ProjectNameOrId: args[0],
				ImportYamlPath:  args[1],
				ParentCmd:       constants.Service,
			})
		},
	}
	cmdImport.Flags().BoolP("help", "h", false, helpText(i18n.ServiceImportHelp))
	return cmdImport
}
