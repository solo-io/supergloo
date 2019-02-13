package consul

// Tested working for consul chart 0.6.0
// be sure to test if adding new versions!

const helmValues = ``
const helmValues2 = `
# Enable Connect for secure communication between nodes
connectInject:
  enabled: {{ .AutoInject.Enabled }}

client:
  enabled: true
  grpc: {{ .AutoInject.Enabled }}

# Install only one Consul server
# This should never be greater than the number of available nodes
server:
  replicas: {{ .NodeCount }}
  bootstrapExpect: {{ .NodeCount }}
  disruptionBudget:
    enabled: true
    maxUnavailable: 0

  # connect will enable Connect on all the servers, initializing a CA
  # for Connect-related connections. Other customizations can be done
  # via the extraConfig setting.
  connect: true
`
