# HOW TO USE: append or source this script from your .zshrc file to clean the temporary kubeconfig file
# set by KUBECONFIG env variable when exiting the shell

trap kubeswitchCleanupHandler EXIT

function kubeswitchCleanupHandler {
 if [ ! -z "$KUBECONFIG" ]
 then
    switchTmpDirectory="$HOME/.kube/.switch_tmp/config"
    if [[ $KUBECONFIG == *"$switchTmpDirectory"* ]]; then
      rm $KUBECONFIG
    fi
 fi
}