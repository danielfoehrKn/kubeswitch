#!/usr/bin/env bash

function switch(){
#  if the executable path is not set, the switcher binary has to be on the path
# this is the case when installing it via homebrew
  DEFAULT_EXECUTABLE_PATH='switcher'
  REPORT_RESPONSE=''
  declare -a opts

  while test $# -gt 0; do
    case "$1" in
    --executable-path)
        EXECUTABLE_PATH=$1
        ;;
    -c | --current)
        opts+=( current-context )
        REPORT_RESPONSE=$1
        ;;
    -u | --unset)
        opts+=( unset-context )
        ;;
    -d)
        opts+=( delete-context )
        ;;
    hooks)
        REPORT_RESPONSE=$1
        ;;
    *)
        opts+=( $1 )
        ;;
    esac
    shift
  done

  if [ -z "$EXECUTABLE_PATH" ]
  then
    EXECUTABLE_PATH=$DEFAULT_EXECUTABLE_PATH
  fi

  RESPONSE=($EXECUTABLE_PATH ${opts[@]})
  if [ $? -ne 0 ]
  then
    return $?
  fi

  if [ -n "$RESPONSE" ]
  then
    if [ -n "$REPORT_RESPONSE" ]
    then
      echo $RESPONSE
    fi

    # first, cleanup old temporary kubeconfig file
    switchTmpDirectory="$HOME/.kube/.switch_tmp/config"
    if [[ -n "$KUBECONFIG" && $KUBECONFIG == *"$switchTmpDirectory"* ]]
    then
      rm -f $KUBECONFIG
    fi

    export KUBECONFIG=$RESPONSE
    currentContext=$(kubectl config current-context)
    echo "switched to context \"$currentContext\"."
  fi
}
