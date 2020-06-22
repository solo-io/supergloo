// NOTE: Generated from ./copyYamlToGoVars.bash
// DO NOT MODIFY

package operator_manifests

const OperatorWithProfile = `apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  namespace: {{ .InstallNamespace }}
  name: istiocontrolplane-{{ .Profile }}
spec:
  profile: {{ .Profile }}
`
