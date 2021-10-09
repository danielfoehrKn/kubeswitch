_kube_contexts()
{
  local curr_arg;
  curr_arg=${COMP_WORDS[COMP_CWORD]}

  # if not the first argument, and the previous one is not a flag (so it is a command)
  if [ "$COMP_CWORD" -gt 1 ]; then

      case ${COMP_WORDS[COMP_CWORD - 1]} in
      alias*)
        arguments="ls
    rm
    --state-directory
    --config-path
    --kubeconfig-name
    --kubeconfig-path
    --no-index
    --store
    --vault-api-address
    --help"
        ;;

      clean*)
        arguments=""
        ;;

      hooks*)
        arguments="ls
    --config-path
    --hook-name
    --run-immediately
    --state-directory
    --help"
        ;;

      list-contexts*)
        arguments="--config-path
    --kubeconfig-name
    --kubeconfig-path
    --no-index
    --state-directory
    --store
    --vault-api-address
    --help"
        ;;

      *)
        arguments=""
        ;;

    esac

    if [[ $arguments != "" ]]; then
      COMPREPLY=( $(compgen -W "$arguments") );
      return
    fi
  fi

  if [[ $curr_arg != --* ]]; then
    contexts=$(switch list-contexts)
  fi

  COMPREPLY=( $(compgen -W "history
  help
  clean
  hooks
  alias
  list-contexts
  --kubeconfig-path
  --no-index
  --debug
  --store
  --kubeconfig-name
  --show-preview
  --vault-api-address
  --executable-path
  --state-directory
  --config-path
  --help
  -c
  --current
  -d
  -u
  --unset
  -
  .
  -v
  version
  $contexts " -- $curr_arg ) );
}

complete -F _kube_contexts switch