package drivers

import (
	"context"

	hclog "github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	"github.com/hashicorp/nomad/plugins/base"
	baseproto "github.com/hashicorp/nomad/plugins/base/proto"
	"github.com/hashicorp/nomad/plugins/drivers/proto"
	"google.golang.org/grpc"
)

// PluginDriver wraps a DriverPlugin and implements go-plugins GRPCPlugin
// interface to expose the the interface over gRPC
type PluginDriver struct {
	plugin.NetRPCUnsupportedPlugin
	impl DriverPlugin
}

func NewDriverPlugin(d DriverPlugin) plugin.GRPCPlugin {
	return &PluginDriver{
		impl: d,
	}
}

func (p *PluginDriver) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterDriverServer(s, &driverPluginServer{
		impl:   p.impl,
		broker: broker,
	})
	return nil
}

func (p *PluginDriver) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &driverPluginClient{
		BasePluginClient: &base.BasePluginClient{
			DoneCtx: ctx,
			Client:  baseproto.NewBasePluginClient(c),
		},
		client:  proto.NewDriverClient(c),
		doneCtx: ctx,
	}, nil
}

// Serve is used to serve a driverplugin
func Serve(d DriverPlugin, logger hclog.Logger) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: base.Handshake,
		Plugins: map[string]plugin.Plugin{
			base.PluginTypeBase:   &base.PluginBase{Impl: d},
			base.PluginTypeDriver: &PluginDriver{impl: d},
		},
		GRPCServer: plugin.DefaultGRPCServer,
		Logger:     logger,
	})
}