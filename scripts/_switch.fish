function __fish_kubeswitch_arg_number -a number
    set -l cmd (commandline -opc)
    test (count $cmd) -eq $number
end

function __fish_kubeswitch_contexts
    set -l cmd (commandline -opc)
    set -l lastArg (string escape -- (commandline -ct))

    # if not the first argument, and the previous one is not a flag (so it is a command)
    if test (count $cmd) -gt 1

      switch "$cmd[-2]"
      case alias
        set -f arguments ls \
          rm \
          --state-directory \
          --config-path \
          --kubeconfig-name \
          --kubeconfig-path \
          --no-index \
          --store \
          --vault-api-address \
          --help

      case clean
        set -f arguments ""

      case hooks
        set -f arguments ls \
          --config-path \
          --hook-name \
          --run-immediately \
          --state-directory \
          --help

      case list-contexts
        set -f arguments --config-path \
          --kubeconfig-name \
          --kubeconfig-path \
          --no-index \
          --state-directory \
          --store \
          --vault-api-address \
          --help

      case '*'
        set -f arguments ""

    end

    if test -n "$arguments"
      printf "%s\n" $arguments
      return
    end
  end

  set -f arguments history \
    help \
    version \
    clean \
    hooks \
    alias \
    list-contexts \
    --kubeconfig-path \
    --no-index \
    --debug \
    --store \
    --kubeconfig-name \
    --show-preview \
    --vault-api-address \
    --executable-path \
    --state-directory \
    --config-path \
    -h --help \
    -c --current \
    -d \
    -u --unset \
    - \
    . \
    -v --version

  if string match -q -v -- '--*' $lastArg
    set -a arguments (kubeswitch list-contexts)
  end

  printf "%s\n" $arguments
end


complete -c kubeswitch -e
complete -f -c kubeswitch
complete -f -x -c kubeswitch -a "(__fish_kubeswitch_contexts)"
