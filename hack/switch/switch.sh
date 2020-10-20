#!/usr/bin/env bash

hookUsage()
{
  echo -e "  --hook-config-path path to the hook configuration file. (default \"$HOME/.kube/switch-config.yaml\")"
  echo -e "  --hook-state-directory path to the state directory. (default \"$HOME/.kube/switch-state\")"
}

usage()
{
   echo "Usage:"

   # usage for `switch hooks`
   if [ -n "$1" ]
  then
    hookUsage
    echo -e "  --hook-name the name of the hook that should be run."
    echo -e "  --run-hooks-immediately run hooks right away. Do not respect the hooks execution configuration. (default \"true\")."
    return
  fi

   echo -e "  --kubeconfig-directory directory containing the kubeconfig files. (default \"$HOME/.kube\")"
   echo -e "  --kubeconfig-name shows kubeconfig files with this name. Accepts wilcard arguments '*' and '?'. (default \"config\")"
   echo -e "  --executable-path path to the 'switcher' executable. If unset tries to use 'switcher' from the path."
   echo -e "  --show-preview if it should show a preview. Preview is sanitized from credentials. (default \"true\")"
   hookUsage
   echo -e "  --help shows available flags."
   echo -e "  clean removes all the temporary kubeconfig files created in the directory \"$HOME/.kube/switch_tmp\"."
}

function switch(){
#  if the executable path is not set, the switcher binary has to be on the path
# this is the case when installing it via homebrew
  DEFAULT_EXECUTABLE_PATH='switcher'

  KUBECONFIG_DIRECTORY=''
  KUBECONFIG_NAME=''
  EXECUTABLE_PATH=''
  SHOW_PREVIEW=''
  CLEAN=''

  # Hooks
  HOOKS=''
  CONFIG_PATH=''
  STATE_DIRECTORY=''
  NAME=''
  RUN_IMMEDIATELY=''

  while test $# -gt 0; do
             case "$1" in
                  --kubeconfig-directory)
                      shift
                      KUBECONFIG_DIRECTORY=$1
                      shift
                      ;;
                  --kubeconfig-name)
                      shift
                      KUBECONFIG_NAME=$1
                      shift
                      ;;
                  --executable-path)
                      shift
                      EXECUTABLE_PATH=$1
                      shift
                      ;;
                  --show-preview)
                      shift
                      SHOW_PREVIEW=$1
                      shift
                      ;;
                  clean)
                      CLEAN=$1
                      shift
                      ;;
                  hooks)
                      HOOKS=$1
                      shift
                      ;;
                  --hook-config-path)
                      shift
                      CONFIG_PATH=$1
                      shift
                      ;;
                  --hook-state-directory)
                      shift
                      STATE_DIRECTORY=$1
                      shift
                      ;;
                  --hook-name)
                      shift
                      # hook name
                      NAME=$1
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
                     usage $HOOKS
                     return
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

  KUBECONFIG_DIRECTORY_FLAG=''
  if [ -n "$KUBECONFIG_DIRECTORY" ]
  then
     KUBECONFIG_DIRECTORY="$KUBECONFIG_DIRECTORY"
     KUBECONFIG_DIRECTORY_FLAG=--kubeconfig-directory
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

  CONFIG_PATH_FLAG=''
  if [ -n "$CONFIG_PATH" ]
  then
     CONFIG_PATH="$CONFIG_PATH"
     CONFIG_PATH_FLAG=--config-path
  fi

  STATE_DIRECTORY_FLAG=''
  if [ -n "$STATE_DIRECTORY" ]
  then
     STATE_DIRECTORY="$STATE_DIRECTORY"
     STATE_DIRECTORY_FLAG=--state-directory
  fi

  if [ -n "$HOOKS" ]
  then
     echo "Running hooks."

     NAME_FLAG=''
     if [ -n "$NAME" ]
     then
        NAME="$NAME"
        NAME_FLAG=--name
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
     $NAME_FLAG ${NAME})

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
  NEW_KUBECONFIG=$($EXECUTABLE_PATH $KUBECONFIG_DIRECTORY_FLAG ${KUBECONFIG_DIRECTORY} $KUBECONFIG_NAME_FLAG ${KUBECONFIG_NAME} $SHOW_PREVIEW_FLAG=${SHOW_PREVIEW})
  if [[ "$?" = "0" ]]; then
      export KUBECONFIG=${NEW_KUBECONFIG}
      currentContext=$(kubectl config current-context)
    echo "switched to context $currentContext"
  fi
}
