package util

import kubeconfigutil "github.com/danielfoehrkn/kubectlSwitch/pkg/util/kubectx_copied"

// GetContextForAlias returns the alias for the given context or an empty string given a map (context -> alias)
func GetContextForAlias(context string, mapping map[string]string) string {
	if value, ok := mapping[context]; ok {
		return value
	}
	if value, ok := mapping[kubeconfigutil.GetContextWithoutFolderPrefix(context)]; ok {
		return value
	}
	return ""
}
