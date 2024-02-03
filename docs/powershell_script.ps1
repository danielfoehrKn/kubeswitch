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
	param (
		[string]$opts
	)

	#You need to have switcher_windows_amd64.exe in your PATH, or you need to change the value of EXECUTABLE_PATH here
	$EXECUTABLE_PATH = "switcher_windows_amd64.exe"

	if (-not $opts) {
	Write-Output "no options provided"
		Write-Output $EXECUTABLE_PATH $opts
		$RESPONSE = & $EXECUTABLE_PATH 
	} 
	else{
	Write-Output "options provided:" $opts
			Write-Output $EXECUTABLE_PATH $opts
		$RESPONSE = & $EXECUTABLE_PATH  $opts
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