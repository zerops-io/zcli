package buildDeploy

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zerops-io/zcli/src/utils/processChecker"
)

func (h *Handler) Deploy(ctx context.Context, config RunConfig) error {
	serviceStack, err := h.checkInputValues(ctx, config)
	if err != nil {
		return err
	}

	fmt.Println(i18n.BuildDeployCreatingPackageStart)

	files, err := h.archClient.FindFilesByRules(config.WorkingDir, config.PathsForPacking)
	if err != nil {
		return err
	}

	reader, writer := io.Pipe()
	defer reader.Close()

	tarErrChan := make(chan error, 1)

	go h.archClient.TarFiles(writer, files, tarErrChan)

	r, err := h.savePackage(config, reader)
	if err != nil {
		return err
	}

	configContent, err := getConfigContent(config)
	if err != nil {
		return err
	}

	appVersion, err := h.createAppVersion(ctx, config, serviceStack)
	if err != nil {
		return err
	}

	if err := h.packageUpload(appVersion, r); err != nil {
		// if an error occurred while packing the app, return that error
		select {
		case err := <-tarErrChan:
			return err
		default:
			return err
		}
	}

	// wait for packing and saving to finish (should already be done after the package upload has finished)
	if tarErr := <-tarErrChan; tarErr != nil {
		return tarErr
	}

	fmt.Println(i18n.BuildDeployDeployingStart)

	deployResponse, err := h.apiGrpcClient.PutAppVersionDeploy(ctx, &zBusinessZeropsApiProtocol.PutAppVersionDeployRequest{
		Id:            appVersion.GetId(),
		ConfigContent: configContent,
	})
	if err := proto.BusinessError(deployResponse, err); err != nil {
		return err
	}

	deployProcessId := deployResponse.GetOutput().GetId()

	err = processChecker.CheckProcess(ctx, deployProcessId, h.apiGrpcClient)
	if err != nil {
		return err
	}

	fmt.Println(constants.Success + i18n.BuildDeploySuccess)

	return nil
}

func getConfigContent(config RunConfig) (*zBusinessZeropsApiProtocol.StringNull, error) {
	workingDir, err := filepath.Abs(config.WorkingDir)
	if err != nil {
		return nil, err
	}

	if config.ZeropsYamlPath != nil {
		workingDir = filepath.Join(workingDir, *config.ZeropsYamlPath)
	}

	zeropsYamlPath := filepath.Join(workingDir, zeropsYamlFileName)

	zeropsYamlStat, err := os.Stat(zeropsYamlPath)
	if err != nil {
		if os.IsNotExist(err) {
			if config.ZeropsYamlPath != nil {
				return nil, errors.New(i18n.BuildDeployZeropsYamlNotFound)
			}
		}
		return nil, nil
	}

	fmt.Printf("%s: %s\n", i18n.BuildDeployZeropsYamlFound, zeropsYamlPath)

	if zeropsYamlStat.Size() == 0 {
		return nil, errors.New(i18n.BuildDeployZeropsYamlEmpty)
	}
	if zeropsYamlStat.Size() > 10*1024 {
		return nil, errors.New(i18n.BuildDeployZeropsYamlTooLarge)
	}

	yamlContent, err := os.ReadFile(zeropsYamlPath)
	if err != nil {
		return nil, err
	}

	return &zBusinessZeropsApiProtocol.StringNull{
		Value: base64.StdEncoding.EncodeToString(yamlContent),
		Valid: true,
	}, nil
}
