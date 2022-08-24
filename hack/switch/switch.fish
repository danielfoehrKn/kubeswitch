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
      case -c or --current
        set -a opts current-context
        set -f REPORT_RESPONSE $i
      case -u or --unset
        set -a opts unset-context
      case -d
        set -a opts delete-context
      case hooks
        set -f REPORT_RESPONSE $i
      case *
        set -a opts $1
    end
  end

  if test -z "$EXECUTABLE_PATH"
    set -f EXECUTABLE_PATH $DEFAULT_EXECUTABLE_PATH
  end

  set -f RESPONSE ($EXECUTABLE_PATH $opts)
  if test $status -ne 0
    return $status
  end

  if test -n "$RESPONSE"
    if test -n "$REPORT_RESPONSE"
      echo $RESPONSE
      return
    end
    
    # first, cleanup old temporary kubeconfig file
    if test -n "$KUBECONFIG"
      set switchTmpDirectory "$HOME/.kube/.switch_tmp/config"
      if string match -q *"$switchTmpDirectory"* -- $KUBECONFIG
        rm -f $KUBECONFIG
      end
    end

    set -x KUBECONFIG $RESPONSE
    set -l currentContext (kubectl config current-context)
    echo "switched to context \"$currentContext\"."
  end
end
