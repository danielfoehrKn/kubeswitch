package switcher

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	shellScript string = `#!/usr/bin/env bash

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
}`

	fishScript string = `#!/usr/bin/env fish

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
      printf "%s\n" $RESPONSE
    end`
)

var (
	initCmd = &cobra.Command{
		Use:                   "init [bash|zsh|fish]",
		Short:                 "generate init and completion script",
		Long:                  "generate and load the init and completion script for switch into the current shell. Use it like this: 'source <(switcher init zsh)'",
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish"},
		Args:                  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Root()
			if setName != "" {
				root.Use = setName
			}
			switch args[0] {
			case "bash":
				fmt.Println(shellScript)
				return root.GenBashCompletion(os.Stdout)
			case "zsh":
				fmt.Println(shellScript)
				return root.GenZshCompletion(os.Stdout)
			case "fish":
				fmt.Println(fishScript)
				return root.GenFishCompletion(os.Stdout, true)
			}
			return fmt.Errorf("unsupported shell type: %s", args[0])
		},
	}
)

func init() {
	rootCommand.AddCommand(initCmd)
}
