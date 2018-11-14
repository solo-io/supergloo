package utils

import (
	"bytes"
	"github.com/pkg/errors"
	"text/template"
)

func render(tmpl *template.Template, data interface{}) (string, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return "", errors.Wrapf(err, "executing template")
	}
	return buf.String(), nil
}

// basic prometheus deployment
func BasicPrometheusDeployment(namespace, name, configmapName string) (string, error) {
	data := struct {
		Namespace, Name, ConfigmapName string
	}{
		Namespace:     namespace,
		Name:          name,
		ConfigmapName: configmapName,
	}
	return render(basicPrometheusDeploymentTemplate, data)
}

var basicPrometheusDeploymentTemplate = template.Must(template.New("").Parse(basicPrometheusDeployment))

const basicPrometheusDeployment = `
# Source: https://raw.githubusercontent.com/bibinwilson/kubernetes-prometheus/master/prometheus-deployment.yaml
# with fixes
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus-server
  template:
    metadata:
      labels:
        app: prometheus-server
    spec:
      containers:
        - name: prometheus
          image: prom/prometheus:v2.2.1
          args:
            - "--config.file=/etc/prometheus/prometheus.yml"
            - "--storage.tsdb.path=/prometheus/"
          ports:
            - containerPort: 9090
          volumeMounts:
            - name: prometheus-config-volume
              mountPath: /etc/prometheus/
            - name: prometheus-storage-volume
              mountPath: /prometheus/
      volumes:
        - name: prometheus-config-volume
          configMap:
            defaultMode: 420
            name: {{ .ConfigmapName }}
  
        - name: prometheus-storage-volume
          emptyDir: {}
`

// basic prometheus service

func BasicPrometheusService(namespace, name string, port uint32) (string, error) {
	data := struct {
		Namespace, Name, ConfigmapName string
		Port                           uint32
	}{
		Namespace: namespace,
		Name:      name,
		Port:      port,
	}
	return render(basicPrometheusServiceTemplate, data)
}

var basicPrometheusServiceTemplate = template.Must(template.New("").Parse(basicPrometheusService))

const basicPrometheusService = `
# source: https://raw.githubusercontent.com/bibinwilson/kubernetes-prometheus/master/prometheus-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
  annotations:
      prometheus.io/scrape: 'true'
      prometheus.io/path:   /
      prometheus.io/port:   '8080'
  
spec:
  selector: 
    app: prometheus-server
  type: NodePort  
  ports:
    - port: 8080
      targetPort: 9090 
      nodePort: {{ .Port }}
`

// istio prometheus deployment
// not currently working
func IstioPrometheusDeployment(namespace, name, configmapName string) (string, error) {
	data := struct {
		Namespace, Name, ConfigmapName string
	}{
		Namespace:     namespace,
		Name:          name,
		ConfigmapName: configmapName,
	}
	return render(istioPrometheusDeploymentTemplate, data)
}

var istioPrometheusDeploymentTemplate = template.Must(template.New("").Parse(istioPrometheusDeployment))

const istioPrometheusDeployment = `
 Source: istio/charts/prometheus/templates/deployment.yaml
# TODO: the original template has service account, roles, etc
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: prometheus
  namespace: istio-system
  labels:
    app: prometheus
    chart: prometheus-1.0.3
    release: istio
    heritage: Tiller
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      labels:
        app: prometheus
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      serviceAccountName: prometheus
      containers:
      - name: prometheus
        image: "docker.io/prom/prometheus:v2.3.1"
        imagePullPolicy: IfNotPresent
        args:
        - '--storage.tsdb.retention=6h'
        - '--config.file=/etc/prometheus/prometheus.yml'
        ports:
        - containerPort: 9090
          name: http
        livenessProbe:
          httpGet:
            path: /-/healthy
            port: 9090
        readinessProbe:
          httpGet:
            path: /-/ready
            port: 9090
        resources:
          requests:
            cpu: 10m

        volumeMounts:
        - name: config-volume
          mountPath: /etc/prometheus
        - mountPath: /etc/istio-certs
          name: istio-certs
      volumes:
      - name: config-volume
        configMap:
          name: prometheus
      - name: istio-certs
        secret:
          defaultMode: 420
          optional: true
          secretName: istio.default
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x
`

// istio prometheus service

func IstioPrometheusService(namespace, name string) (string, error) {
	data := struct {
		Namespace, Name, ConfigmapName string
	}{
		Namespace: namespace,
		Name:      name,
	}
	return render(istioPrometheusServiceTemplate, data)
}

var istioPrometheusServiceTemplate = template.Must(template.New("").Parse(istioPrometheusService))

const istioPrometheusService = `
# Source: istio/charts/prometheus/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: prometheus
  namespace: istio-system
  annotations:
    prometheus.io/scrape: 'true'
  labels:
    name: prometheus
spec:
  selector:
    app: prometheus
  ports:
  - name: http-prometheus
    protocol: TCP
    port: 9090
`
