# Return exit code 1 if anything in `pkg` imports from `github.com/solo-io/skv2/contrib`

if go list  -f '{{ join .Deps "\n" }}'  github.com/solo-io/gloo-mesh/pkg/... | grep -x "github.com/solo-io/skv2/contrib"; then
  echo "Error: found imports to 'github.com/solo-io/skv2/contrib' in pkg"
  exit 1
fi
