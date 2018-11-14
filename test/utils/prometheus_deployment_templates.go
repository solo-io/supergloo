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


// prometheus deployment
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
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
spec:
  replicas: 1
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



func BasicPrometheusService(namespace, name string) (string, error) {
	data := struct {
		Namespace, Name, ConfigmapName string
	}{
		Namespace:     namespace,
		Name:          name,
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
      nodePort: 30000
`
