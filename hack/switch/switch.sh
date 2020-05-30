#!/usr/bin/env bash

usage()
{
   echo "Usage:"
   echo -e "  --kubeconfig-directory directory containing the kubeconfig files. Default is $HOME/.kube/switch"
   echo -e "  --kubeconfig-name only shows kubeconfig files with exactly this name. Defaults to 'config'."
   echo -e "  --executable-path path to the 'switch' executable. If unset tries to use 'switch' from the path."
   echo -e "  --show-preview if it should show a preview. Preview is sanitized from credentials. Defaults to true."
   echo -e "  --help shows available flags."
}

switch(){
#  if the executable path is not set, the switcher binary has to be on the path
# this is the case when installing it via homebrew
  DEFAULT_EXECUTABLE_PATH='switcher'

  KUBECONFIG_DIRECTORY=''
  KUBECONFIG_NAME=''
  EXECUTABLE_PATH=''
  SHOW_PREVIEW=''

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
                  --help)
                     usage
                     return
                     ;;
                  -h)
                     usage
                     return
                     ;;
                  *)
                     usage
                     return
                     ;;
            esac
    done

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

  SHOW_PREVIEW_FLAG=''
  if [ -n "$SHOW_PREVIEW" ]
  then
     SHOW_PREVIEW="$SHOW_PREVIEW"
     SHOW_PREVIEW_FLAG=--show-preview
  fi

  if [ -z "$EXECUTABLE_PATH" ]
  then
     EXECUTABLE_PATH=$DEFAULT_EXECUTABLE_PATH
  fi

  # execute golang binary handing over all the flags
  NEW_KUBECONFIG=$($EXECUTABLE_PATH $KUBECONFIG_DIRECTORY_FLAG ${KUBECONFIG_DIRECTORY} $KUBECONFIG_NAME_FLAG ${KUBECONFIG_NAME} $SHOW_PREVIEW_FLAG ${SHOW_PREVIEW})
  if [[ "$?" = "0" ]]; then
      export KUBECONFIG=${NEW_KUBECONFIG}
      currentContext=$(kubectl config current-context)
    echo "switched to context $currentContext"
  fi
}
