package generator

import (
	"text/template"
)

var BasicServiceManifestTemplate = template.Must(
	template.New("basic_service_manifest").
		Funcs(GenBuildFuncs).
		Parse(`
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: {{ kebab .AppGroup }}
    {{ kebab .AppGroup }}: {{ kebab .AppName }}
  annotations:
    "helm.sh/hook": "pre-install"
    "helm.sh/hook-weight": "5"
  name: {{ kebab .AppName }}
  namespace: {{ wrap " $.Release.Namespace "}}
---
{{wrap "- if .Values.global.rbac.create"}}
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: {{ kebab .AppName }}-role-binding
  labels:
    app: {{ kebab .AppGroup }}
    {{ kebab .AppGroup }}: {{ kebab .AppName }}
subjects:
  - kind: ServiceAccount
    name: {{ kebab .AppName }}
    namespace: {{ wrap " .Release.Namespace "}}
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
  {{ wrap "- end "}}
---
{{wrap "- if not .Values.meshDiscovery.disabled "}}
{{wrap "- $image := .Values.meshDiscovery.deployment.image "}}
{{wrap "- if .Values.global  "}}
{{wrap "- $image = merge .Values.meshDiscovery.deployment.image .Values.global.image "}}
{{wrap "- end "}}
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: {{ kebab .AppGroup }}
    {{ kebab .AppGroup }}: {{ kebab .AppName }}
  name: {{ kebab .AppName }}
  namespace: {{wrap " .Release.Namespace "}}
spec:
  replicas: {{wrap " .Values.meshDiscovery.deployment.replicas "}}
  selector:
    matchLabels:
      {{ kebab .AppGroup }}: {{ kebab .AppName }}
  template:
    metadata:
      labels:
        {{ kebab .AppGroup }}: {{ kebab .AppName }}
      {{wrap "- if .Values.meshDiscovery.deployment.stats "}}
      annotations:
        prometheus.io/path: /metrics
        prometheus.io/port: "9091"
        prometheus.io/scrape: "true"
      {{wrap "- end"}}
    spec:
      serviceAccountName: {{ kebab .AppName }}
      volumes:
        - name: aws-cred
          secret:
            secretName: aws-cred
      containers:
      - image: {{wrap "template \"gloo.image\" $image"}}
        imagePullPolicy: {{wrap " $image.pullPolicy "}}
        name: {{ kebab .AppName }}
        ## If using EKS, create a secret called "aws-cred" in the .Release.Namespace namespace
        ## and include the following key-value pair:
        ## "credentials": <base64-encoded aws credentials, as can be gotten from: cat ~/.aws/credentials | base64>
        volumeMounts:
          - mountPath: "/root/.aws/"
            name: aws-cred
            readOnly: true
{{wrap "- if .Values.meshDiscovery.deployment.resources "}}
        resources:
{{wrap " toYaml .Values.meshDiscovery.deployment.resources | indent 10"}}
{{wrap "- end"}}
        env:
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
        {{wrap "- if .Values.meshDiscovery.deployment.stats "}}
          - name: START_STATS_SERVER
            value: "true"
        {{wrap "- end"}}
{{wrap "- end "}}
`))
