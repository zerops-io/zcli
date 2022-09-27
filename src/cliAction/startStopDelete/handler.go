package startStopDelete

import (
	"context"

	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zeropsio/zcli/src/utils/httpClient"
	"github.com/zeropsio/zcli/src/utils/sdkConfig"
)

type Config struct {
}

type Method func(ctx context.Context, projectId string, serviceId string) (string, error)

type CmdType struct {
	Start   string // message for cmd start
	Finish  string // message for cmd end
	Execute Method
}

type RunConfig struct {
	ProjectNameOrId string
	ServiceName     string
	Confirm         bool
	ParentCmd       constants.ParentCmd
	CmdData         CmdType
}

func (c *RunConfig) getCmdProps() (string, string, Method) {
	cd := c.CmdData
	return cd.Start, cd.Finish, cd.Execute
}

type Handler struct {
	config        Config
	httpClient    *httpClient.Handler
	apiGrpcClient zBusinessZeropsApiProtocol.ZBusinessZeropsApiProtocolClient
	sdkConfig     sdkConfig.Config
}

func New(config Config, httpClient *httpClient.Handler, apiGrpcClient zBusinessZeropsApiProtocol.ZBusinessZeropsApiProtocolClient, sdkConfig sdkConfig.Config) *Handler {
	return &Handler{
		config:        config,
		httpClient:    httpClient,
		apiGrpcClient: apiGrpcClient,
		sdkConfig:     sdkConfig,
	}
}
