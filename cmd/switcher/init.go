// Copyright 2021 The Kubeswitch authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package switcher

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	shellScript string = `
has_prefix() { case $2 in "$1"*) true;; *) false;; esac; }
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

  if [ -z "$EXECUTABLE_PATH" ]; then
	EXECUTABLE_PATH="$DEFAULT_EXECUTABLE_PATH"
  fi

  RESPONSE="$($EXECUTABLE_PATH "${opts[@]}")"
  if [ $? -ne 0 -o -z "$RESPONSE" ]; then
	printf "%s\n" "$RESPONSE"
	return $?
  fi

  # switcher returns a response that contains a kubeconfig path with a prefix "__ " to be able to
  # distinguish it from other responses which just need to write to STDOUT
  prefix="__ "
  if ! has_prefix "$prefix" "$RESPONSE" ; then
	  printf "%s\n" "$RESPONSE"
	  return
  fi

  # remove prefix
  RESPONSE=${RESPONSE#"$prefix"}

  #the response form the switcher binary is "kubeconfig_path,selected_context"
  remainder="$RESPONSE"
  KUBECONFIG_PATH="${remainder%%,*}"; remainder="${remainder#*,}"
  SELECTED_CONTEXT="${remainder%%,*}"; remainder="${remainder#*,}"

  if [ -z ${KUBECONFIG_PATH+x} ]; then
	# KUBECONFIG_PATH is not set
	printf "%s\n" "$RESPONSE"
	return
  fi

  if [ -z ${SELECTED_CONTEXT+x} ]; then
	# SELECTED_CONTEXT is not set
	printf "%s\n" "$RESPONSE"
	return
  fi

  # cleanup old temporary kubeconfig file
  local switchTmpDirectory="$HOME/.kube/.switch_tmp/config"
  if [[ -n "$KUBECONFIG" && "$KUBECONFIG" == *"$switchTmpDirectory"* ]]
  then
	\rm -f "$KUBECONFIG"
  fi

  export KUBECONFIG="$KUBECONFIG_PATH"
  printf "switched to context %s\n" "$SELECTED_CONTEXT"
}`

	fishScript string = `
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
	  command rm -f "$KUBECONFIG"
	end

	set -gx KUBECONFIG "$KUBECONFIG_PATH"
	printf "switched to context %s\n" "$SELECTED_CONTEXT"
	return
  end
  printf "%s\n" $RESPONSE
    end`

	powershellScript string = `
function has_prefix {
	param (
		[string]$prefix,
		[string]$string
	)

	if ($string.StartsWith($prefix)) {
		return $true
	} else {
		return $false
	}
}

function kubeswitch {

	#You need to have switcher_windows_amd64.exe in your PATH, or you need to change the value of EXECUTABLE_PATH here
	$EXECUTABLE_PATH = "switcher_windows_amd64.exe"

	if (-not $args) {
	Write-Output "no options provided"
		Write-Output $EXECUTABLE_PATH $args
		$RESPONSE = & $EXECUTABLE_PATH 
	} 
	else{
	Write-Output "options provided:" $args
			Write-Output $EXECUTABLE_PATH $args
		$RESPONSE = & $EXECUTABLE_PATH  $args
	}

	if ($LASTEXITCODE -ne 0 -or -not $RESPONSE) {
		Write-Output $RESPONSE
		return $LASTEXITCODE
	}

	# switcher returns a response that contains a kubeconfig path with a prefix "__ " to be able to
	# distinguish it from other responses which just need to write to STDOUT
	$prefix = "__ "
	if (-not (has_prefix $prefix $RESPONSE)) {
		Write-Output $RESPONSE
		return
	}


	$RESPONSE = $RESPONSE -replace $prefix, ""
	Write-Output $RESPONSE
	$remainder = $RESPONSE
	Write-Output $remainder
	Write-Output $remainder.split(",")[0]
	Write-Output $remainder.split(",")[1]
	$KUBECONFIG_PATH = $remainder.split(",")[0]
	$KUBECONFIG_PATH = $KUBECONFIG_PATH -replace '\\', '/'
	$KUBECONFIG_PATH = $KUBECONFIG_PATH -replace "C:", ""
	Write-Output $KUBECONFIG_PATH
	$SELECTED_CONTEXT = $remainder.split(",")[1]

	if (-not $KUBECONFIG_PATH) { 
		Write-Output $RESPONSE
		return
	}

	if (-not $SELECTED_CONTEXT) {
		Write-Output $RESPONSE
		return
	}

	$switchTmpDirectory = "$env:USERPROFILE\.kube\.switch_tmp\config"
	if ($env:KUBECONFIG -and $env:KUBECONFIG -like "*$switchTmpDirectory*") {
		Remove-Item -Path $env:KUBECONFIG -Force
	}

	$env:KUBECONFIG = $KUBECONFIG_PATH
	Write-Output "switched to context $SELECTED_CONTEXT"
}

#Env variable HOME doesn't exist on windows, we create it from USERPROFILE
$Env:HOME = $Env:USERPROFILE
`
)

var (
	initCmd = &cobra.Command{
		Use:                   "init [bash|zsh|fish|powershell]",
		Short:                 "generate init and completion script",
		Long:                  "generate and load the init and completion script for switch into the current shell. Use it like this: 'source <(switcher init zsh)'",
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Root()
			if setName != "" {
				root.Use = setName
			}
			switch args[0] {
			case "bash":
				// same shell script as zsh, but different bash completion
				fmt.Println(shellScript)
				return root.GenBashCompletion(os.Stdout)
			case "zsh":
				fmt.Println(shellScript)
				return root.GenZshCompletion(os.Stdout)
			case "fish":
				fmt.Println(fishScript)
				return root.GenFishCompletion(os.Stdout, true)
			case "powershell":
				fmt.Println(powershellScript)
				return root.GenPowerShellCompletion(os.Stdout)
			}
			return fmt.Errorf("unsupported shell type: %s", args[0])
		},
	}
)

func init() {
	rootCommand.AddCommand(initCmd)
}
