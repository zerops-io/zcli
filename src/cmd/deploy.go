package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/zeropsio/zcli/src/cliAction/buildDeploy"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zeropsio/zcli/src/utils/archiveClient"
	"github.com/zeropsio/zcli/src/utils/httpClient"
	"github.com/zeropsio/zcli/src/utils/sdkConfig"
)

func deployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "deploy projectNameOrId serviceName pathToFileOrDir [pathToFileOrDir] [flags]",
		Short:        i18n.CmdDeployDesc,
		Long:         i18n.CmdDeployDesc + "\n\n" + i18n.DeployDescLong + "\n\n" + i18n.DeployHintPush,
		SilenceUsage: true,
		Args:         MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
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

			arch := archiveClient.New(archiveClient.Config{})

			return buildDeploy.New(
				buildDeploy.Config{},
				client,
				arch,
				apiGrpcClient,
				sdkConfig.Config{Token: token, RegionUrl: reg.RestApiAddress},
			).Deploy(ctx, buildDeploy.RunConfig{
				ArchiveFilePath:  params.GetString(cmd, "archiveFilePath"),
				WorkingDir:       params.GetString(cmd, "workingDir"),
				VersionName:      params.GetString(cmd, "versionName"),
				ZeropsYamlPath:   params.GetString(cmd, "zeropsYamlPath"),
				ProjectNameOrId:  args[0],
				ServiceStackName: args[1],
				PathsForPacking:  args[2:],
			})
		},
	}

	params.RegisterString(cmd, "workingDir", "./", i18n.BuildWorkingDir)
	params.RegisterString(cmd, "archiveFilePath", "", i18n.BuildArchiveFilePath)
	params.RegisterString(cmd, "versionName", "", i18n.BuildVersionName)
	params.RegisterString(cmd, "zeropsYamlPath", "", i18n.ZeropsYamlLocation)

	cmd.Flags().BoolP("help", "h", false, helpText(i18n.DeployHelp))

	return cmd
}
