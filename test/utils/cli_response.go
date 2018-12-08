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

const RoutingRulesArgs = `
	{{.RuleType}} test-rule -s --mesh {{.MeshName}} --namespace supergloo-system
	--destinations {{.Upstream}} --sources {{.Upstream}}
`

const HeaderManipulationArgs = `
    --header.request.append header-request1,text --header.request.remove header-request2 
    --header.response.append header-response1,text --header.response.remove header-response2
`
const FaultInjectionArgs = `
	--fault.abort.message 404
	--fault.abort.percent 50
	--fault.abort.type http
	--fault.delay.percent 50
	--fault.delay.type fixed
	--fault.delay.value 1s
`
const RetriesArgs = `
	--route.retry.attempt 3
	--route.retry.timeout 1s
`
const TimeoutArgs = `
    --route.timeout 1s2ms3ns
`
const TrafficShiftingArgs = ` --traffic.upstreams {{.Upstream}} --traffic.weights 1 `
const MirrorArgs = ` --mirror {{.Upstream}} `
const CorsArgs = `
	--cors.allow.credentials
	--cors.allow.headers *
	--cors.allow.methods get
	--cors.allow.origin *
	--cors.expose.headers authorization
	--cors.maxage 10s
`
