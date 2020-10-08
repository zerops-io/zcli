package cmd

import (
	"context"

	"github.com/zerops-io/zcli/src/grpcDaemonClientFactory"

	"github.com/zerops-io/zcli/src/i18n"

	"github.com/zerops-io/zcli/src/cliAction/stopVpn"

	"github.com/spf13/cobra"
)

func vpnStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "stop",
		Short:        i18n.CmdVpnStop,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			regSignals(cancel)

			daemonClient, daemonCloseFunc, err := grpcDaemonClientFactory.New().CreateClient(ctx)
			if err != nil {
				return err
			}
			defer daemonCloseFunc()

			return stopVpn.New(
				stopVpn.Config{},
				daemonClient,
			).Run(ctx, stopVpn.RunConfig{})
		},
	}

	return cmd
}
