package login

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/proto"
	"github.com/zeropsio/zcli/src/proto/daemon"
	"github.com/zeropsio/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zeropsio/zcli/src/utils/httpClient"
)

type Config struct {
	RestApiAddress string
	GrpcApiAddress string
}

type RunConfig struct {
	ZeropsEmail    string
	ZeropsPassword string
	ZeropsToken    string
}

type Handler struct {
	config               Config
	storage              *cliStorage.Handler
	httpClient           *httpClient.Handler
	grpcApiClientFactory *zBusinessZeropsApiProtocol.Handler
}

func New(
	config Config,
	storage *cliStorage.Handler,
	httpClient *httpClient.Handler,
	grpcApiClientFactory *zBusinessZeropsApiProtocol.Handler,
) *Handler {
	return &Handler{
		config:               config,
		storage:              storage,
		httpClient:           httpClient,
		grpcApiClientFactory: grpcApiClientFactory,
	}
}

func (h *Handler) Run(ctx context.Context, runConfig RunConfig) error {

	if runConfig.ZeropsPassword == "" &&
		runConfig.ZeropsEmail == "" &&
		runConfig.ZeropsToken == "" {
		return errors.New(i18n.LoginParamsMissing)
	}

	var err error
	if runConfig.ZeropsToken != "" {
		err = h.loginWithToken(ctx, runConfig.ZeropsToken)
	} else {
		err = h.loginWithPassword(ctx, runConfig.ZeropsEmail, runConfig.ZeropsPassword)
	}
	if err != nil {
		return err
	}

	daemonClient, closeFunc, err := daemon.CreateClient(ctx)
	if err != nil {
		return err
	}
	defer closeFunc()

	response, err := daemonClient.StopVpn(ctx, &daemon.StopVpnRequest{})
	daemonInstalled, err := proto.DaemonError(err)
	if err != nil {
		return err
	}

	if daemonInstalled && response.GetTunnelState() == daemon.TunnelState_TUNNEL_SET_INACTIVE {
		fmt.Println(i18n.LoginVpnClosed)
	}

	fmt.Println(i18n.LoginSuccess)
	return nil
}

func (h *Handler) loginWithPassword(_ context.Context, login, password string) error {
	loginData, err := json.Marshal(struct {
		Email    string
		Password string
	}{
		Email:    login,
		Password: password,
	})
	if err != nil {
		return err
	}

	loginResponse, err := h.httpClient.Post(h.config.RestApiAddress+"/api/rest/public/auth/login", loginData)
	if err != nil {
		return err
	}

	var loginResponseObject struct {
		Auth struct {
			AccessToken string
		}
	}

	if loginResponse.StatusCode < http.StatusBadRequest {
		err := json.Unmarshal(loginResponse.Body, &loginResponseObject)
		if err != nil {
			return err
		}
	} else {
		return parseRestApiError(loginResponse.Body)
	}

	cliResponse, err := h.httpClient.Post(
		h.config.RestApiAddress+"/api/rest/public/user-token",
		nil,
		httpClient.BearerAuthorization(loginResponseObject.Auth.AccessToken),
	)
	if err != nil {
		return err
	}

	if cliResponse.StatusCode >= http.StatusBadRequest {
		return parseRestApiError(cliResponse.Body)
	}

	var tokenResponseObject struct {
		Token string `json:"token"`
	}
	err = json.Unmarshal(cliResponse.Body, &tokenResponseObject)
	if err != nil {
		return err
	}

	h.storage.Update(func(data cliStorage.Data) cliStorage.Data {
		data.Token = tokenResponseObject.Token
		return data
	})

	return nil
}

func (h *Handler) loginWithToken(ctx context.Context, token string) error {
	grpcApiClient, closeFunc, err := h.grpcApiClientFactory.CreateClient(ctx, h.config.GrpcApiAddress, token)
	if err != nil {
		return err
	}
	defer closeFunc()

	resp, err := grpcApiClient.GetUserInfo(ctx, &zBusinessZeropsApiProtocol.GetUserInfoRequest{})

	if err := proto.BusinessError(resp, err); err != nil {
		if proto.IsUnauthenticated(err) {
			return i18n.AddHintChangeRegion(err)
		}
		return err
	}

	h.storage.Update(func(data cliStorage.Data) cliStorage.Data {
		data.Token = token
		return data
	})

	return nil
}
