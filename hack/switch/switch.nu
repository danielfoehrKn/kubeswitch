const default_executable_path = "switcher"

export def --env --wrapped main [
    ...args
]: nothing -> nothing {
  let result = (switch-inner ...$args)
  if $result.exit_code? != null {
    return
  }

  $env.KUBECONFIG = $result.kubeconfig
  print $"switched to context ($result.context)"
}

def --wrapped switch-inner [
    --executable-path: path
    ...args
] {
  let opts = []

  let executable_path = ($executable_path | default $default_executable_path)

  let response = (do {
     ^$executable_path ...$args
  } | complete)
  if ($response.exit_code != 0 or ($response.stdout | str trim) == "") {
    print $response.stdout
    return {exit_code: $response.exit_code}
  }

  let resp_text = ($response.stdout | str trim)
  let prefix = "__ "

  if not ($resp_text | str starts-with $prefix) {
    print $resp_text
    return {exit_code: 0}
  }

  # Remove prefix
  let resp_text = ($resp_text | str replace $prefix "")

  # Parse out kubeconfig path and context
  let parts = ($resp_text | split row ",")
  if ($parts | length) != 2 {
    error make {
      msg: $"got an unexpected output from switcher: ($resp_text)"
    }
  }
  let kubeconfig_path = ($parts | get 0)
  let selected_context = ($parts | get 1)

  if $kubeconfig_path == "" {
    error make { msg: $resp_text }
  }

  if $selected_context == "" {
    error make { msg: $resp_text }
  }

  # Cleanup old temporary kubeconfig file
  let switch_tmp_directory = $"($env.HOME)/.kube/.switch_tmp/config"
  if ($env.KUBECONFIG? != null and ($env.KUBECONFIG | str contains $switch_tmp_directory)) {
    rm $env.KUBECONFIG
  }

  return {
    kubeconfig: $kubeconfig_path
    context: $selected_context
  }
}
