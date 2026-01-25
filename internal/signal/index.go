package signal

import (
	"context"
	"errors"
	"fmt"

	plugrpc "github.com/sailorhq/plug/sdk/proto"
	v1 "github.com/sailorhq/sailor/pkg/core/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	FailurePolicyFail   = "Fail"
	FailurePolicyIgnore = "Ignore"
)

type ErrorInfo struct {
	Plug string
	Err  error
}

type SailorTx interface {
	Load(plugs []v1.RxSetting)
	FireDeploy(req *plugrpc.DeployRequest) error
	Cleanup()
}

// PlugManager implements IPlugger interface which contributes to the
// main plugin system of Sailor. This entity ensures that all the plugins
// are loaded successfully with formal verification from plugin side.
//
// When the Plugger fires the `OnBoot` signal to the plug, it is expected
// by the plug to provide a DoneResponse with .Ok as true. Which means
// all the dependencies of the plug was initialized successfully and
// Sailor can provide signals the plug as per the lifecycle.
//
// NOTE: It is required by all plugins to acknowledge this signal or els
// the signals will not be passed on to the plugin ever.
type PlugManager struct {
	// ActivePlugCtlMap contains only the active plugs which is verified with OnBoot signal
	ActivePlugCtlMap map[string]plugrpc.RxClient
	// ActiveSettingMap contains only active plug settings
	ActiveSettingMap map[string]v1.RxSetting
	Log              *zap.Logger
}

func NewPlugManager(log *zap.Logger) *PlugManager {
	return &PlugManager{
		ActivePlugCtlMap: make(map[string]plugrpc.RxClient),
		ActiveSettingMap: make(map[string]v1.RxSetting),
		Log:              log,
	}
}

func (p *PlugManager) Load(plugs []v1.RxSetting) {
	for _, rx := range plugs {
		noFunctionalityMsg := "any functionality from this plug will not work"
		if !rx.Enabled {
			p.Log.Warn("a plug was not enabled, any functionality from this plug will not work",
				zap.String("rx_name", rx.Name))
			continue
		}

		// initialize plugin client
		conn, err := grpc.NewClient(
			"localhost:"+rx.Port,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			p.Log.Error("plug does not implement SailorHook",
				zap.String("plug_name", rx.Name),
				zap.Error(err))
			continue
		}
		client := plugrpc.NewRxClient(conn)

		bootResp, err := client.OnBootSignal(context.TODO(), &plugrpc.BootRequest{})
		if !bootResp.Ok {
			p.Log.Error("plug OnBoot has failed",
				zap.String("rx_name", rx.Name),
				zap.String("err", bootResp.Message))

			// cleanup the process that we just started because we might not need
			// a failed initialized process running and hogging CPU
			p.Log.Info("closing connection on plug",
				zap.String("rx_name", rx.Name))
			conn.Close()

			p.Log.Warn(noFunctionalityMsg)
			continue
		}

		p.Log.Info(fmt.Sprintf("plug %s has been locked and loaded", rx.Name), zap.String("rx_port", rx.Port))

		p.ActivePlugCtlMap[rx.Name] = client
		p.ActiveSettingMap[rx.Name] = rx // TODO: check if we can use reference here
	}
}

func (p *PlugManager) FireDeploy(req *plugrpc.DeployRequest) error {
	// we loop through all the active plugs
	for name, ctl := range p.ActivePlugCtlMap {
		resp, err := ctl.OnDeploySignal(context.TODO(), req)
		if err != nil && p.ActiveSettingMap[name].FailurePolicy.Signal.OnDeploy == FailurePolicyFail {
			return err
		}
		if !resp.Ack.Ok && p.ActiveSettingMap[name].FailurePolicy.Signal.OnDeploy == FailurePolicyFail {
			p.Log.Error("plug OnDeploy failed with error",
				zap.String("rx_name", name),
				zap.String("err", resp.Ack.Message))
			return errors.New(resp.Ack.Message)
		}
	}

	return nil
}
