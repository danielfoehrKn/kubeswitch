#!/usr/bin/env fish

function kubeswitch
#  if the executable path is not set, the switcher binary has to be on the path
# this is the case when installing it via homebrew
  set -f DEFAULT_EXECUTABLE_PATH 'switcher'
  set -f REPORT_RESPONSE
  set -f opts

  for i in $argv
    switch "$i"
      case --executable-path
        set -f EXECUTABLE_PATH $i
      case completion
        set -a opts $i --cmd kubeswitch
      case '*'
        set -a opts $i
    end
  end

  if test -z "$EXECUTABLE_PATH"
    set -f EXECUTABLE_PATH $DEFAULT_EXECUTABLE_PATH
  end

  set -f RESULT 0
  set -f RESPONSE ($EXECUTABLE_PATH $opts; or set RESULT $status | string split0)
  if test $RESULT -ne 0; or test -z "$RESPONSE"
    printf "%s\n" $RESPONSE
    return $RESULT
  end

  # switcher returns a response that contains a kubeconfig path with a prefix "__ " to be able to
  # distinguish it from other responses which just need to write to STDOUT
  if string match -q "__ *" -- "$RESPONSE"
    # remove the prefix
    set -l RESPONSE (string replace --regex "__ " "" "$RESPONSE")
    set split_info (string split , "$RESPONSE")

    if set -q split_info[1]
        set KUBECONFIG_PATH $split_info[1]
    else
        # kubeconfig path is not set, simply return the response
        printf "%s\n" $RESPONSE
        return
    end

    if set -q split_info[2]
        set SELECTED_CONTEXT $split_info[2]
    else
        # context is not set, simply return the response
        printf "%s\n" $RESPONSE
        return
    end

    if test ! -e "$KUBECONFIG_PATH"
      echo "ERROR: \"$KUBECONFIG_PATH\" does not exist"
      return 1
    end

    set -l switchTmpDirectory "$HOME/.kube/.switch_tmp/config"
    if test -n "$KUBECONFIG"; and string match -q "*$switchTmpDirectory*" -- "$KUBECONFIG"
      \rm -f "$KUBECONFIG"
    end

    set -gx KUBECONFIG "$KUBECONFIG_PATH"
    printf "switched to context %s\n" "$SELECTED_CONTEXT"
    return
  end
  printf "%s\n" $RESPONSE
end
