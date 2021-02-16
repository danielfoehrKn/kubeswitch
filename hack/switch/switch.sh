#!/usr/bin/env bash

hooksUsage() {
echo '
Run configured hooks

Usage:
  switch hooks [flags]

Flags:
      --config-path string        path on the local filesystem to the configuration file. (default "~/.kube/switch-config.yaml")
  -h, --help                      help for hooks
      --hook-name string          the name of the hook that should be run.
      --run-immediately           run hooks right away. Do not respect the hooks execution configuration. (default true)
      --state-directory string    path to the state directory. (default "~/.kube/switch-state")
'
}

switchUsage() {
echo '
The kubectx for operators.

Usage:
  switch [flags]
  switch [command]

Available Commands:
  <context-name>  Switch to context name provided as first argument
  help            Help about any command
  history         Switch to a previous context from the history (short: h)
  clean           Cleans all temporary kubeconfig files
  hooks           Runs configured hooks

Flags:
      --config-path string         path on the local filesystem to the configuration file. (default "~/.kube/switch-config.yaml")
  -h, --help                       help for switch
      --kubeconfig-name string     only shows Kubeconfig files with this name. Accepts wilcard arguments "*" and "?". Defaults to "config". (default "config")
      --kubeconfig-path string     path to be recursively searched for Kubeconfig files. Can be a file or directory on the local filesystem or a path in Vault. (default "~/.kube/config")
      --show-preview               show preview of the selected Kubeconfig. Possibly makes sense to disable when using vault as the Kubeconfig store to prevent excessive requests against the API. (default true)
      --state-directory string     path to the local directory used for storing internal state. (default "~/.kube/switch-state")
      --store string               the backing store to be searched for Kubeconfig files. Can be either "filesystem" or "vault" (default "filesystem")
      --vault-api-address string   the API address of the Vault store. Overrides the default "vaultAPIAddress" field in the SwitchConfig. This flag is overridden by the environment variable "VAULT_ADDR".

Use "switch [command] --help" for more information about a command.
'
}

unknownCommand() {
  echo "error: unknown command \"$1\" for "switch"
Run 'switch --help' for usage."
}

usage()
{
   # usage for `switch hooks`
   if [ -n "$1" ]
  then
    hooksUsage
    return
  fi

  switchUsage
}


function switch(){
#  if the executable path is not set, the switcher binary has to be on the path
# this is the case when installing it via homebrew
  DEFAULT_EXECUTABLE_PATH='switcher'

  KUBECONFIG_PATH=''
  STORE=''
  KUBECONFIG_NAME=''
  SHOW_PREVIEW=''
  CONFIG_PATH=''
  VAULT_API_ADDRESS=''
  EXECUTABLE_PATH=''
  CLEAN=''
  SET_CONTEXT=''
  HISTORY=''

  # Hooks
  HOOKS=''
  STATE_DIRECTORY=''
  HOOK_NAME=''
  RUN_IMMEDIATELY=''

  while test $# -gt 0; do
             case "$1" in
                  --kubeconfig-path)
                      shift
                      KUBECONFIG_PATH=$1
                      shift
                      ;;
                  --store)
                      shift
                      STORE=$1
                      shift
                      ;;
                  --kubeconfig-name)
                      shift
                      KUBECONFIG_NAME=$1
                      shift
                      ;;
                  --show-preview)
                      shift
                      SHOW_PREVIEW=$1
                      shift
                      ;;
                  --vault-api-address)
                      shift
                      VAULT_API_ADDRESS=$1
                      shift
                      ;;
                  --executable-path)
                      shift
                      EXECUTABLE_PATH=$1
                      shift
                      ;;
                  clean)
                      CLEAN=$1
                      shift
                      ;;
                  h)
                      HISTORY=$1
                      shift
                      ;;
                  history)
                      HISTORY=$1
                      shift
                      ;;
                  hooks)
                      HOOKS=$1
                      shift
                      ;;
                  --state-directory)
                      shift
                      STATE_DIRECTORY=$1
                      shift
                      ;;
                  --config-path)
                      shift
                      CONFIG_PATH=$1
                      shift
                      ;;
                  --hook-name)
                      shift
                      HOOK_NAME=$1
                      shift
                      ;;
                  --run-hooks-immediately)
                      shift
                      RUN_IMMEDIATELY=$1
                      shift
                      ;;
                  --help)
                     usage $HOOKS
                     return
                     ;;
                  -h)
                     usage $HOOKS
                     return
                     ;;
                  *)
                     SET_CONTEXT=$1
                     shift
                     ;;
            esac
    done

  if [ -z "$EXECUTABLE_PATH" ]
  then
     EXECUTABLE_PATH=$DEFAULT_EXECUTABLE_PATH
  fi

  if [ -n "$CLEAN" ]
  then
     $EXECUTABLE_PATH clean
     return
  fi

  if [ -n "$HISTORY" ]
  then
     NEW_KUBECONFIG=$($EXECUTABLE_PATH history)
     setKubeconfigEnvironmentVariable $NEW_KUBECONFIG
     return
  fi

  KUBECONFIG_PATH_FLAG=''
  if [ -n "$KUBECONFIG_PATH" ]
  then
     KUBECONFIG_PATH="$KUBECONFIG_PATH"
     KUBECONFIG_PATH_FLAG=--kubeconfig-path
  fi

  STORE_FLAG=''
  if [ -n "$STORE" ]
  then
     STORE="$STORE"
     STORE_FLAG=--store
  fi

  KUBECONFIG_NAME_FLAG=''
  if [ -n "$KUBECONFIG_NAME" ]
  then
     KUBECONFIG_NAME="$KUBECONFIG_NAME"
     KUBECONFIG_NAME_FLAG=--kubeconfig-name
  fi

  SHOW_PREVIEW_FLAG=--show-preview
  if [ -n "$SHOW_PREVIEW" ]
  then
     SHOW_PREVIEW="$SHOW_PREVIEW"
  else
     SHOW_PREVIEW="true"
  fi

  VAULT_API_ADDRESS_FLAG=''
  if [ -n "$VAULT_API_ADDRESS" ]
  then
     VAULT_API_ADDRESS="$VAULT_API_ADDRESS"
     VAULT_API_ADDRESS_FLAG=--vault-api-address
  fi

  STATE_DIRECTORY_FLAG=''
  if [ -n "$STATE_DIRECTORY" ]
  then
     STATE_DIRECTORY="$STATE_DIRECTORY"
     STATE_DIRECTORY_FLAG=--state-directory
  fi

  CONFIG_PATH_FLAG=''
  if [ -n "$CONFIG_PATH" ]
  then
     CONFIG_PATH="$CONFIG_PATH"
     CONFIG_PATH_FLAG=--config-path
  fi

  if [ -n "$SET_CONTEXT" ]
  then
     SET_CONTEXT="$SET_CONTEXT"
  fi

  if [ -n "$HOOKS" ]
  then
     echo "Running hooks."

     HOOK_NAME_FLAG=''
     if [ -n "$HOOK_NAME" ]
     then
        HOOK_NAME="$HOOK_NAME"
        HOOK_NAME_FLAG=--hook-name
     fi

     RUN_IMMEDIATELY_FLAG=--run-immediately
     if [ -n "$RUN_IMMEDIATELY" ]
     then
        RUN_IMMEDIATELY="$RUN_IMMEDIATELY"
     else
        RUN_IMMEDIATELY="true"
     fi

     RESPONSE=$($EXECUTABLE_PATH hooks \
     $RUN_IMMEDIATELY_FLAG=${RUN_IMMEDIATELY} \
     $CONFIG_PATH_FLAG ${CONFIG_PATH} \
     $STATE_DIRECTORY_FLAG ${STATE_DIRECTORY} \
     $HOOK_NAME_FLAG ${HOOK_NAME})

      if [ -n "$RESPONSE" ]
      then
         echo $RESPONSE
      fi
     return
  fi

  # always run hooks command with --run-immediately=false
  $EXECUTABLE_PATH hooks \
     --run-immediately=false \
     $CONFIG_PATH_FLAG ${CONFIG_PATH} \
     $STATE_DIRECTORY_FLAG ${STATE_DIRECTORY}

  # execute golang binary handing over all the flags
  NEW_KUBECONFIG=$($EXECUTABLE_PATH \
  $SET_CONTEXT \
  $KUBECONFIG_PATH_FLAG ${KUBECONFIG_PATH} \
  $STORE_FLAG ${STORE} \
  $KUBECONFIG_NAME_FLAG ${KUBECONFIG_NAME} \
  $SHOW_PREVIEW_FLAG=${SHOW_PREVIEW} \
  $VAULT_API_ADDRESS_FLAG ${VAULT_API_ADDRESS} \
  $STATE_DIRECTORY_FLAG ${STATE_DIRECTORY} \
  $CONFIG_PATH_FLAG ${CONFIG_PATH})

  setKubeconfigEnvironmentVariable $NEW_KUBECONFIG
}

function setKubeconfigEnvironmentVariable() {
  if [[ "$?" = "0" ]]; then
    export KUBECONFIG=${NEW_KUBECONFIG}
    currentContext=$(kubectl config current-context)
    echo "switched to context $currentContext"
  fi
}
