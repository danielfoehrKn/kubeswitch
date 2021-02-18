_kube_contexts()
{
  local curr_arg;
  curr_arg=${COMP_WORDS[COMP_CWORD]}

  if [[ $curr_arg != --* ]]; then
    contexts=$(switch list-contexts)
  fi

  COMPREPLY=( $(compgen -W "history
  help
  clean
  hooks
  --kubeconfig-path
  --store
  --kubeconfig-name
  --show-preview
  --vault-api-address
  --executable-path
  --state-directory
  --config-path
  --hook-name
  --run-hooks-immediately
  --help
  $contexts " -- $curr_arg ) );
}

complete -F _kube_contexts switch