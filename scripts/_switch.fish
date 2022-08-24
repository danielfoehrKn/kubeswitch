function __fish_kubeswitch_arg_number -a number
    set -l cmd (commandline -opc)
    test (count $cmd) -eq $number
end

complete -f -c kubeswitch
complete -f -x -c kubeswitch -n '__fish_kubeswitch_arg_number 1' -a "(kubeswitch list-contexts)"
complete -f -x -c kubeswitch -n '__fish_kubeswitch_arg_number 1' -a "-" -d "switch to the previous namespace in this context"
