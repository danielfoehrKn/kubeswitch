package plugins

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	kubeconfigstorev1 "github.com/danielfoehrkn/kubeswitch/pkg/store/plugins/kubeconfigstore/v1"
	storetypes "github.com/danielfoehrkn/kubeswitch/pkg/store/types"
)

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "KUBESWITCH_PLUGIN",
	MagicCookieValue: "kubeswitch",
}

type Store interface {
	GetID(ctx context.Context) (string, error)
	GetContextPrefix(ctx context.Context, path string) (string, error)
	VerifyKubeconfigPaths(ctx context.Context) error
	StartSearch(ctx context.Context, channel chan storetypes.SearchResult)
	GetKubeconfigForPath(ctx context.Context, path string, tags map[string]string) ([]byte, error)
}

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"store": &StorePlugin{},
}

// StorePlugin is the implementation of plugin.Plugin so we can serve/consume this.
type StorePlugin struct {
	plugin.NetRPCUnsupportedPlugin

	Impl Store
}

func (p *StorePlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	kubeconfigstorev1.RegisterKubeconfigStoreServiceServer(s, &GRPCServer{Impl: p.Impl})

	return nil
}

func (p *StorePlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: kubeconfigstorev1.NewKubeconfigStoreServiceClient(c)}, nil
}

var _ plugin.GRPCPlugin = &StorePlugin{}
