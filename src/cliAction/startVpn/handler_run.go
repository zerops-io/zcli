package startVpn

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/proto/daemon"
	"github.com/zerops-io/zcli/src/utils/projectService"

	"github.com/zerops-io/zcli/src/daemonInstaller"

	"github.com/peterh/liner"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils"
)

func (h *Handler) Run(ctx context.Context, config RunConfig) error {

	userInfoResponse, err := h.apiGrpcClient.GetUserInfo(ctx, &business.GetUserInfoRequest{})
	if err := proto.BusinessError(userInfoResponse, err); err != nil {
		return err
	}
	userId := userInfoResponse.GetOutput().GetId()

	project, err := projectService.GetProject(ctx, h.apiGrpcClient, config.ProjectName)
	if err != nil {
		return err
	}

	err = h.tryStartVpn(ctx, project, userId, config)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) tryStartVpn(ctx context.Context, project *business.Project, userId string, config RunConfig) error {

	zeropsDaemonClient, closeFn, err := daemon.CreateClient(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	response, err := zeropsDaemonClient.StartVpn(ctx, &daemon.StartVpnRequest{
		ApiAddress:       h.config.GrpcApiAddress,
		VpnAddress:       h.config.VpnAddress,
		ProjectId:        project.GetId(),
		Token:            config.Token,
		Mtu:              config.Mtu,
		UserId:           userId,
		CaCertificateUrl: config.CaCertificateUrl,
	})
	daemonInstalled, err := proto.DaemonError(err)
	if err != nil {
		return err
	}
	if !daemonInstalled {
		fmt.Println(i18n.VpnStartDaemonIsUnavailable)

		line := liner.NewLiner()
		defer line.Close()
		line.SetCtrlCAborts(true)

		fmt.Println(i18n.VpnStartInstallDaemonPrompt)
		for {
			if answer, err := line.Prompt("y/n "); err == nil {
				if answer == "n" {
					return errors.New(i18n.VpnStartTerminatedByUser)
				} else if answer == "y" {
					err := h.daemonInstaller.Install()

					if errors.Is(err, daemonInstaller.ErrElevatedPrivileges) {
						return nil
					}

					if err != nil {
						return err
					}
					fmt.Println(i18n.DaemonInstallSuccess)

					// let's wait for daemon start
					time.Sleep(3 * time.Second)
					return h.tryStartVpn(ctx, project, userId, config)
				} else {
					fmt.Println(i18n.VpnStartUserIsUnableToWriteYorN)
					continue
				}
			} else if err == liner.ErrPromptAborted {
				return errors.New(i18n.VpnStartTerminatedByUser)
			} else {
				return err
			}
		}
	}

	utils.PrintVpnStatus(response.GetVpnStatus())
	return nil
}
