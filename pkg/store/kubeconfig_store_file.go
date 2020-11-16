package store

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/karrick/godirwalk"
	"github.com/sirupsen/logrus"

	"github.com/danielfoehrkn/kubectlSwitch/types"
)

func (s *FilesystemStore) GetKind() types.StoreKind {
	return types.StoreKindFilesystem
}

func (s *FilesystemStore) DiscoverPaths(log *logrus.Entry, channel chan PathDiscoveryResult) {
	var kubeconfigPaths []string

	if err := godirwalk.Walk(s.KubeconfigPath, &godirwalk.Options{
		Callback: func(osPathname string, _ *godirwalk.Dirent) error {
			fileName := filepath.Base(osPathname)
			matched, err := filepath.Match(s.KubeconfigName, fileName)
			if err != nil {
				return err
			}
			if matched {
				kubeconfigPaths = append(kubeconfigPaths, osPathname)
				channel <- PathDiscoveryResult{
					KubeconfigPath: osPathname,
					Error:          nil,
				}
			}
			return nil
		},
		Unsorted: false, // (optional) set true for faster yet non-deterministic enumeration
	}); err != nil {
		channel <- PathDiscoveryResult{
			KubeconfigPath: "",
			Error:          fmt.Errorf("failed to find kubeconfig files in directory: %v", err),
		}
	}
}

func (s *FilesystemStore) GetKubeconfigForPath(log *logrus.Entry, path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func (s *FilesystemStore) CheckRootPath() error {
	if _, err := os.Stat(s.KubeconfigPath); os.IsNotExist(err) {
		return fmt.Errorf("the kubeconfig directory %q does not exist", s.KubeconfigPath)
	}
	return nil
}
