package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/zeropsio/zcli/src/cmd/scope"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/nettools"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zerops-go/errorCode"
)

func ExecuteCmd() error {
	return cmdBuilder.ExecuteRootCmd(rootCmd())
}

func rootCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("zcli").
		SetHelpTemplate(getRootTemplate()).
		SilenceError(true).
		AddChildrenCmd(loginCmd()).
		AddChildrenCmd(versionCmd()).
		AddChildrenCmd(scopeCmd()).
		AddChildrenCmd(projectCmd()).
		AddChildrenCmd(serviceCmd()).
		AddChildrenCmd(vpnCmd()).
		AddChildrenCmd(statusShowDebugLogsCmd()).
		AddChildrenCmd(servicePushCmd()).
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			body := &uxBlock.TableBody{}

			body.AddStringsRow(i18n.T(i18n.StatusInfoLoggedUser), "-")

			guestInfoPart(body)

			cmdData.UxBlocks.Table(body)

			// print the default command help
			cmdData.PrintHelp()

			return nil
		}).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			body := &uxBlock.TableBody{}

			var loggedUser string
			if info, err := cmdData.RestApiClient.GetUserInfo(ctx); err != nil {
				loggedUser = err.Error()
			} else {
				if infoOutput, err := info.Output(); err != nil {
					loggedUser = err.Error()
				} else {
					loggedUser = fmt.Sprintf("%s <%s>", infoOutput.FullName, infoOutput.Email)
				}
			}

			body.AddStringsRow(i18n.T(i18n.StatusInfoLoggedUser), loggedUser)

			guestInfoPart(body)

			if cmdData.CliStorage.Data().ScopeProjectId.Filled() {
				// project scope is set
				projectId, _ := cmdData.CliStorage.Data().ScopeProjectId.Get()
				project, err := repository.GetProjectById(ctx, cmdData.RestApiClient, projectId)
				if err != nil {
					if errorsx.Check(err, errorsx.CheckErrorCode(errorCode.ProjectNotFound)) {
						err := scope.ProjectScopeReset(cmdData)
						if err != nil {
							return err
						}
					} else {
						body.AddStringsRow(i18n.T(i18n.ScopedProject), err.Error())
					}
				} else {
					body.AddStringsRow(i18n.T(i18n.ScopedProject), fmt.Sprintf("%s [%s]", project.Name.String(), project.ID.Native()))
				}
			}

			pingCtx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()

			if err := nettools.Ping(pingCtx, vpnCheckAddress); err != nil {
				body.AddStringsRow(i18n.T(i18n.StatusInfoVpnStatus), i18n.T(i18n.VpnCheckingConnectionIsNotActive))
			} else {
				body.AddStringsRow(i18n.T(i18n.StatusInfoVpnStatus), i18n.T(i18n.VpnCheckingConnectionIsActive))
			}

			cmdData.UxBlocks.Table(body)

			// print the default command help
			cmdData.PrintHelp()

			return nil
		})
}

func guestInfoPart(tableBody *uxBlock.TableBody) {
	cliDataFilePath, _, err := constants.CliDataFilePath()
	if err != nil {
		cliDataFilePath = err.Error()
	}
	tableBody.AddStringsRow(i18n.T(i18n.StatusInfoCliDataFilePath), cliDataFilePath)

	logFilePath, _, err := constants.LogFilePath()
	if err != nil {
		logFilePath = err.Error()
	}
	tableBody.AddStringsRow(i18n.T(i18n.StatusInfoLogFilePath), logFilePath)

	wgConfigFilePath, _, err := constants.WgConfigFilePath()
	if err != nil {
		wgConfigFilePath = err.Error()
	}
	tableBody.AddStringsRow(i18n.T(i18n.StatusInfoWgConfigFilePath), wgConfigFilePath)
}

func getRootTemplate() string {
	return styles.CobraSectionColor().SetString("Usage:").String() + `{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

` + styles.CobraSectionColor().SetString("Aliases:").String() + `
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

` + styles.CobraSectionColor().SetString("Examples:").String() + `
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

` + styles.CobraSectionColor().SetString("Available Commands:").String() + `{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  ` + styles.CobraItemNameColor().SetString("{{rpad .Name .NamePadding }}").String() + ` {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  ` + styles.CobraItemNameColor().SetString("{{rpad .Name .NamePadding }}").String() + ` {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

` + styles.CobraSectionColor().SetString("Additional Commands:").String() + `{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  ` + styles.CobraItemNameColor().SetString("{{rpad .Name .NamePadding }}").String() + ` {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

` + styles.CobraSectionColor().SetString("Flags:").String() + `
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

` + styles.CobraSectionColor().SetString("Global Flags:").String() + `
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

` + styles.CobraSectionColor().SetString("Additional help topics:").String() + `{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  ` + styles.CobraItemNameColor().SetString("{{rpad .CommandPath .CommandPathPadding}}").String() + ` {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

` + styles.CobraSectionColor().SetString("Global Env Variables:").String() + `
  ` + styles.CobraItemNameColor().SetString(constants.CliLogFilePathEnvVar).String() + `     ` + i18n.T(i18n.CliLogFilePathEnvVar) + `
  ` + styles.CobraItemNameColor().SetString(constants.CliDataFilePathEnvVar).String() + `    ` + i18n.T(i18n.CliDataFilePathEnvVar) + `
  ` + styles.CobraItemNameColor().SetString(constants.CliTerminalMode).String() + `     ` + i18n.T(i18n.CliTerminalModeEnvVar) + `

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
}
