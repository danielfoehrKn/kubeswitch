package plugins

import (
	"context"
	"fmt"

	kubeconfigstorev1 "github.com/danielfoehrkn/kubeswitch/pkg/store/plugins/kubeconfigstore/v1"
	storetypes "github.com/danielfoehrkn/kubeswitch/pkg/store/types"
)

type GRPCClient struct {
	Store
	client kubeconfigstorev1.KubeconfigStoreServiceClient
}

func (m *GRPCClient) GetID(ctx context.Context) (string, error) {
	resp, err := m.client.GetID(ctx, &kubeconfigstorev1.GetIDRequest{})
	if err != nil {
		return "", err
	}
	return resp.Id, nil
}

func (m *GRPCClient) GetContextPrefix(ctx context.Context, path string) (string, error) {
	resp, err := m.client.GetContextPrefix(ctx, &kubeconfigstorev1.GetContextPrefixRequest{Path: path})
	if err != nil {
		return "", err
	}
	return resp.Prefix, nil
}

func (m *GRPCClient) VerifyKubeconfigPaths(ctx context.Context) error {
	_, err := m.client.VerifyKubeconfigPaths(ctx, &kubeconfigstorev1.VerifyKubeconfigPathsRequest{})
	return err
}

func (m *GRPCClient) StartSearch(ctx context.Context, channel chan storetypes.SearchResult) {
	stream, err := m.client.StartSearch(ctx, &kubeconfigstorev1.StartSearchRequest{})
	if err != nil {
		channel <- storetypes.SearchResult{
			KubeconfigPath: "",
			Error:          err,
		}
		return
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			close(channel)
			return
		}

		channel <- storetypes.SearchResult{
			KubeconfigPath: resp.KubeconfigPath,
			Tags:           resp.Tags,
			Error:          err,
		}
	}
}

func (m *GRPCClient) GetKubeconfigForPath(ctx context.Context, path string, tags map[string]string) ([]byte, error) {
	resp, err := m.client.GetKubeconfigForPath(ctx, &kubeconfigstorev1.GetKubeconfigForPathRequest{Path: path})
	if err != nil {
		return nil, err
	}

	return resp.Kubeconfig, nil
}

type GRPCServer struct {
	// This is the real implementation
	Impl Store
}

func (m *GRPCServer) GetID(
	ctx context.Context,
	req *kubeconfigstorev1.GetIDRequest,
) (*kubeconfigstorev1.GetIDResponse, error) {
	v, err := m.Impl.GetID(ctx)
	return &kubeconfigstorev1.GetIDResponse{Id: v}, err
}

func (m *GRPCServer) GetContextPrefix(
	ctx context.Context,
	req *kubeconfigstorev1.GetContextPrefixRequest,
) (*kubeconfigstorev1.GetContextPrefixResponse, error) {
	v, err := m.Impl.GetContextPrefix(ctx, req.Path)
	return &kubeconfigstorev1.GetContextPrefixResponse{Prefix: v}, err
}

func (m *GRPCServer) VerifyKubeconfigPaths(
	ctx context.Context,
	req *kubeconfigstorev1.VerifyKubeconfigPathsRequest,
) (*kubeconfigstorev1.VerifyKubeconfigPathsResponse, error) {
	err := m.Impl.VerifyKubeconfigPaths(ctx)
	return &kubeconfigstorev1.VerifyKubeconfigPathsResponse{}, err
}

func (m *GRPCServer) StartSearch(
	req *kubeconfigstorev1.StartSearchRequest,
	stream kubeconfigstorev1.KubeconfigStoreService_StartSearchServer,
) error {
	ch := make(chan storetypes.SearchResult)

	if stream == nil {
		return fmt.Errorf("stream is nil")
	}

	ctx := stream.Context()

	go m.Impl.StartSearch(ctx, ch)
	for v := range ch {
		if err := stream.Send(&kubeconfigstorev1.StartSearchResponse{KubeconfigPath: v.KubeconfigPath, Tags: v.Tags}); err != nil {
			return err
		}
	}
	return nil
}

func (m *GRPCServer) GetKubeconfigForPath(
	ctx context.Context,
	req *kubeconfigstorev1.GetKubeconfigForPathRequest,
) (*kubeconfigstorev1.GetKubeconfigForPathResponse, error) {
	v, err := m.Impl.GetKubeconfigForPath(ctx, req.Path, req.Tags)
	return &kubeconfigstorev1.GetKubeconfigForPathResponse{Kubeconfig: v}, err
}
