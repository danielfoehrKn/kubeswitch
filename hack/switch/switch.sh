#!/usr/bin/env bash

function switch(){
#  if the executable path is not set, the switcher binary has to be on the path
# this is the case when installing it via homebrew

  local DEFAULT_EXECUTABLE_PATH="switcher"
  declare -a opts

  while test $# -gt 0; do
    case "$1" in
    --executable-path)
        EXECUTABLE_PATH="$1"
        ;;
    completion)
        opts+=("$1" --cmd switch)
        ;;
    *)
        opts+=( "$1" )
        ;;
    esac
    shift
  done

  if [ -z "$EXECUTABLE_PATH" ]
  then
    EXECUTABLE_PATH="$DEFAULT_EXECUTABLE_PATH"
  fi

  RESPONSE="$($EXECUTABLE_PATH "${opts[@]}")"
  if [ $? -ne 0 -o -z "$RESPONSE" ]
  then
    printf "%s\n" "$RESPONSE"
    return $?
  fi

  local trim_left="switched to context \""
  local trim_right="\"."
  if [[ "$RESPONSE" == "$trim_left"*"$trim_right" ]]
  then
    local new_config="${RESPONSE#$trim_left}"
    new_config="${new_config%$trim_right}"

    if [ ! -e "$new_config" ]
    then
      echo "ERROR: \"$new_config\" does not exist"
      return 1
    fi

    # cleanup old temporary kubeconfig file
    local switchTmpDirectory="$HOME/.kube/.switch_tmp/config"
    if [[ -n "$KUBECONFIG" && "$KUBECONFIG" == *"$switchTmpDirectory"* ]]
    then
      rm -f "$KUBECONFIG"
    fi

    export KUBECONFIG="$new_config"
  fi
  printf "%s\n" "$RESPONSE"
}
