package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/proto/daemon"

	"github.com/zeropsio/zcli/src/cliAction/removeDaemon"
	"github.com/zeropsio/zcli/src/cliAction/stopVpn"
	"github.com/zeropsio/zcli/src/daemonInstaller"

	"github.com/zeropsio/zcli/src/i18n"

	"github.com/spf13/cobra"
)

func daemonRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "remove",
		Short:        i18n.CmdDaemonRemove,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			regSignals(cancel)

			daemonClient, daemonCloseFunc, err := daemon.CreateClient(ctx)
			if err != nil {
				return err
			}
			defer daemonCloseFunc()

			installer, err := daemonInstaller.New(daemonInstaller.Config{})
			if err != nil {
				return err
			}

			stopVpn := stopVpn.New(
				stopVpn.Config{},
				daemonClient,
			)

			return removeDaemon.New(
				removeDaemon.Config{},
				installer,
				stopVpn,
			).
				Run(ctx, removeDaemon.RunConfig{})
		},
	}

	cmd.PersistentFlags().BoolP("help", "h", false, helpText(i18n.DaemonRemoveHelp))
	return cmd
}
