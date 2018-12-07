package utils

const RootCommandResponse = `supergloo configures resources used by Supergloo server.
	Find more information at https://solo.io

Usage:
  supergloo [command]

Available Commands:
  add-route           Define new route
  config              Configure mesh resources
  cors                Configure cors policy parameters
  create              Create a resource from stdin
  fault-injection     Stress test your mesh with faults
  fortify-ingress     Configure ingress security parameters
  get                 Display one or many supergloo resources
  header-manipulation Configure header manipulation parameters
  help                Help about any command
  init                Initialize supergloo
  install             Install a mesh
  mirror              Configure mirror parameters
  mtls                set mTLS status
  policy              Apply a policy
  retries             Configure retry parameters
  timeout             Configure timeout parameters
  traffic-shifting    Configure traffic shifting parameters
  uninstall           Uninstall a mesh

Flags:
  -f, --filename string   file input
  -h, --help              help for supergloo
  -s, --static            disable interactive mode
      --version           version for supergloo

Use "supergloo [command] --help" for more information about a command.
`
