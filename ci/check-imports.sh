# Ensure that nothing in pkg imports from codegen/groups

if grep -r 'github.com/solo-io/gloo-mesh/codegen/groups' pkg; then
  echo "Found imports to 'github.com/solo-io/gloo-mesh/codegen/groups' in /pkg"
  exit 1
fi
