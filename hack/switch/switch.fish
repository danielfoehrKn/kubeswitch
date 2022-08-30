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
    printf "%s\n" "$RESPONSE"
    return $RESULT
  end

  set -l trim_left "switched to context \""
  set -l trim_right "\"."
  if string match -q "$trim_left*$trim_right" -- "$RESPONSE"
    set -l new_config (string replace -r "$trim_left(.*)$trim_right\$" '$1' -- "$RESPONSE")

    if test ! -e "$new_config"
      echo "ERROR: \"$new_config\" does not exist"
      return 1
    end

    set -l switchTmpDirectory "$HOME/.kube/.switch_tmp/config"
    if test -n "$KUBECONFIG"; and string match -q "*$switchTmpDirectory*" -- "$KUBECONFIG"
      rm -f "$KUBECONFIG"
    end

    set -gx KUBECONFIG "$new_config"
  end
  printf "%s\n" "$RESPONSE"
end
