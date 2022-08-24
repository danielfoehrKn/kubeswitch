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
      case alias \
        or clean \
        or current-context \
        or delete-context \
        or gardener \
        or help \
        or history or h \
        or hooks \
        or list-contexts \
        or namespace or ns \
        or unset-context \
        or version \
        or -h or --help

        set -f REPORT_RESPONSE $i
        set -a opts $i
      case '*'
        set -a opts $i
    end
  end

  if test -z "$EXECUTABLE_PATH"
    set -f EXECUTABLE_PATH $DEFAULT_EXECUTABLE_PATH
  end

  set -f RESULT 0
  set -f RESPONSE ($EXECUTABLE_PATH $opts; or set RESULT $status | string split0)
  if test $RESULT -ne 0
    printf "%s\n" $RESPONSE
    return $RESULT
  end

  if test -n "$RESPONSE"
    if test -n "$REPORT_RESPONSE"
      printf "%s\n" $RESPONSE
      return
    end
    
    # first, cleanup old temporary kubeconfig file
    set -l switchTmpDirectory "$HOME/.kube/.switch_tmp/config"
    if test -n "$KUBECONFIG"; and string match -q "*$switchTmpDirectory*" -- $KUBECONFIG
      rm -f $KUBECONFIG
    end

    if test ! -e "$RESPONSE"
      echo "ERROR: \"$RESPONSE\" does not exist"
      return 1
    end
    set -x KUBECONFIG "$RESPONSE"
    set -l currentContext (kubectl config current-context)
    echo "switched to context \"$currentContext\"."
  end
end
