package shared_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/supergloo/pkg/install/shared"
	"github.com/solo-io/supergloo/test/utils"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Parse", func() {
	It("works", func() {
		out, err := ParseKubeManifest(utils.Linkerd1Yaml)
		Expect(err).NotTo(HaveOccurred())
		list := KubeObjectList{
			&v1.ConfigMap{
				TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "l5d-config",
					Namespace: "linkerd",
				},
				Data: map[string]string{
					"config.yaml": "admin:\n  ip: 0.0.0.0\n  port: 9990\n\n# Namers provide Linkerd with service discovery information.  To use a\n# namer, you reference it in the dtab by its prefix.  We define 4 namers:\n# * /io.l5d.k8s gets the address of the target app\n# * /io.l5d.k8s.http gets the address of the http-incoming Linkerd router on the target app's node\n# * /io.l5d.k8s.h2 gets the address of the h2-incoming Linkerd router on the target app's node\n# * /io.l5d.k8s.grpc gets the address of the grpc-incoming Linkerd router on the target app's node\nnamers:\n- kind: io.l5d.k8s\n- kind: io.l5d.k8s\n  prefix: /io.l5d.k8s.http\n  transformers:\n    # The daemonset transformer replaces the address of the target app with\n    # the address of the http-incoming router of the Linkerd daemonset pod\n    # on the target app's node.\n  - kind: io.l5d.k8s.daemonset\n    namespace: linkerd\n    port: http-incoming\n    service: l5d\n    # hostNetwork: true # Uncomment if using host networking (eg for CNI)\n- kind: io.l5d.k8s\n  prefix: /io.l5d.k8s.h2\n  transformers:\n    # The daemonset transformer replaces the address of the target app with\n    # the address of the h2-incoming router of the Linkerd daemonset pod\n    # on the target app's node.\n  - kind: io.l5d.k8s.daemonset\n    namespace: linkerd\n    port: h2-incoming\n    service: l5d\n    # hostNetwork: true # Uncomment if using host networking (eg for CNI)\n- kind: io.l5d.k8s\n  prefix: /io.l5d.k8s.grpc\n  transformers:\n    # The daemonset transformer replaces the address of the target app with\n    # the address of the grpc-incoming router of the Linkerd daemonset pod\n    # on the target app's node.\n  - kind: io.l5d.k8s.daemonset\n    namespace: linkerd\n    port: grpc-incoming\n    service: l5d\n    # hostNetwork: true # Uncomment if using host networking (eg for CNI)\n- kind: io.l5d.rewrite\n  prefix: /portNsSvcToK8s\n  pattern: \"/{port}/{ns}/{svc}\"\n  name: \"/k8s/{ns}/{port}/{svc}\"\n\n# Telemeters export metrics and tracing data about Linkerd, the services it\n# connects to, and the requests it processes.\ntelemetry:\n- kind: io.l5d.prometheus # Expose Prometheus style metrics on :9990/admin/metrics/prometheus\n- kind: io.l5d.recentRequests\n  sampleRate: 0.25 # Tune this sample rate before going to production\n# - kind: io.l5d.zipkin # Uncomment to enable exporting of zipkin traces\n#   host: zipkin-collector.default.svc.cluster.local # Zipkin collector address\n#   port: 9410\n#   sampleRate: 1.0 # Set to a lower sample rate depending on your traffic volume\n\n# Usage is used for anonymized usage reporting.  You can set the orgId to\n# identify your organization or set `enabled: false` to disable entirely.\nusage:\n  orgId: linkerd-examples-servicemesh\n\n# Routers define how Linkerd actually handles traffic.  Each router listens\n# for requests, applies routing rules to those requests, and proxies them\n# to the appropriate destinations.  Each router is protocol specific.\n# For each protocol (HTTP, HTTP/2, gRPC) we define an outgoing router and\n# an incoming router.  The application is expected to send traffic to the\n# outgoing router which proxies it to the incoming router of the Linkerd\n# running on the target service's node.  The incoming router then proxies\n# the request to the target application itself.  We also define HTTP and\n# HTTP/2 ingress routers which act as Ingress Controllers and route based\n# on the Ingress resource.\nrouters:\n- label: http-outgoing\n  protocol: http\n  servers:\n  - port: 4140\n    ip: 0.0.0.0\n  # This dtab looks up service names in k8s and falls back to DNS if they're\n  # not found (e.g. for external services). It accepts names of the form\n  # \"service\" and \"service.namespace\", defaulting the namespace to\n  # \"default\". For DNS lookups, it uses port 80 if unspecified. Note that\n  # dtab rules are read bottom to top. To see this in action, on the Linkerd\n  # administrative dashboard, click on the \"dtab\" tab, select \"http-outgoing\"\n  # from the dropdown, and enter a service name like \"a.b\". (Or click on the\n  # \"requests\" tab to see recent traffic through the system and how it was\n  # resolved.)\n  dtab: |\n    /ph  => /$/io.buoyant.rinet ;                     # /ph/80/google.com -> /$/io.buoyant.rinet/80/google.com\n    /svc => /ph/80 ;                                  # /svc/google.com -> /ph/80/google.com\n    /svc => /$/io.buoyant.porthostPfx/ph ;            # /svc/google.com:80 -> /ph/80/google.com\n    /k8s => /#/io.l5d.k8s.http ;                      # /k8s/default/http/foo -> /#/io.l5d.k8s.http/default/http/foo\n    /portNsSvc => /#/portNsSvcToK8s ;                 # /portNsSvc/http/default/foo -> /k8s/default/http/foo\n    /host => /portNsSvc/http/default ;                # /host/foo -> /portNsSvc/http/default/foo\n    /host => /portNsSvc/http ;                        # /host/default/foo -> /portNsSvc/http/default/foo\n    /svc => /$/io.buoyant.http.domainToPathPfx/host ; # /svc/foo.default -> /host/default/foo\n  client:\n    kind: io.l5d.static\n    configs:\n    # Use HTTPS if sending to port 443\n    - prefix: \"/$/io.buoyant.rinet/443/{service}\"\n      tls:\n        commonName: \"{service}\"\n\n- label: http-incoming\n  protocol: http\n  servers:\n  - port: 4141\n    ip: 0.0.0.0\n  interpreter:\n    kind: default\n    transformers:\n    - kind: io.l5d.k8s.localnode\n      # hostNetwork: true # Uncomment if using host networking (eg for CNI)\n  dtab: |\n    /k8s => /#/io.l5d.k8s ;                           # /k8s/default/http/foo -> /#/io.l5d.k8s/default/http/foo\n    /portNsSvc => /#/portNsSvcToK8s ;                 # /portNsSvc/http/default/foo -> /k8s/default/http/foo\n    /host => /portNsSvc/http/default ;                # /host/foo -> /portNsSvc/http/default/foo\n    /host => /portNsSvc/http ;                        # /host/default/foo -> /portNsSvc/http/default/foo\n    /svc => /$/io.buoyant.http.domainToPathPfx/host ; # /svc/foo.default -> /host/default/foo\n\n- label: h2-outgoing\n  protocol: h2\n  servers:\n  - port: 4240\n    ip: 0.0.0.0\n  dtab: |\n    /ph  => /$/io.buoyant.rinet ;                       # /ph/80/google.com -> /$/io.buoyant.rinet/80/google.com\n    /svc => /ph/80 ;                                    # /svc/google.com -> /ph/80/google.com\n    /svc => /$/io.buoyant.porthostPfx/ph ;              # /svc/google.com:80 -> /ph/80/google.com\n    /k8s => /#/io.l5d.k8s.h2 ;                          # /k8s/default/h2/foo -> /#/io.l5d.k8s.h2/default/h2/foo\n    /portNsSvc => /#/portNsSvcToK8s ;                   # /portNsSvc/h2/default/foo -> /k8s/default/h2/foo\n    /host => /portNsSvc/h2/default ;                    # /host/foo -> /portNsSvc/h2/default/foo\n    /host => /portNsSvc/h2 ;                            # /host/default/foo -> /portNsSvc/h2/default/foo\n    /svc => /$/io.buoyant.http.domainToPathPfx/host ;   # /svc/foo.default -> /host/default/foo\n  client:\n    kind: io.l5d.static\n    configs:\n    # Use HTTPS if sending to port 443\n    - prefix: \"/$/io.buoyant.rinet/443/{service}\"\n      tls:\n        commonName: \"{service}\"\n\n- label: h2-incoming\n  protocol: h2\n  servers:\n  - port: 4241\n    ip: 0.0.0.0\n  interpreter:\n    kind: default\n    transformers:\n    - kind: io.l5d.k8s.localnode\n      # hostNetwork: true # Uncomment if using host networking (eg for CNI)\n  dtab: |\n    /k8s => /#/io.l5d.k8s ;                             # /k8s/default/h2/foo -> /#/io.l5d.k8s/default/h2/foo\n    /portNsSvc => /#/portNsSvcToK8s ;                   # /portNsSvc/h2/default/foo -> /k8s/default/h2/foo\n    /host => /portNsSvc/h2/default ;                    # /host/foo -> /portNsSvc/h2/default/foo\n    /host => /portNsSvc/h2 ;                            # /host/default/foo -> /portNsSvc/h2/default/foo\n    /svc => /$/io.buoyant.http.domainToPathPfx/host ;   # /svc/foo.default -> /host/default/foo\n\n- label: grpc-outgoing\n  protocol: h2\n  servers:\n  - port: 4340\n    ip: 0.0.0.0\n  identifier:\n    kind: io.l5d.header.path\n    segments: 1\n  dtab: |\n    /hp  => /$/inet ;                                # /hp/linkerd.io/8888 -> /$/inet/linkerd.io/8888\n    /svc => /$/io.buoyant.hostportPfx/hp ;           # /svc/linkerd.io:8888 -> /hp/linkerd.io/8888\n    /srv => /#/io.l5d.k8s.grpc/default/grpc;         # /srv/service/package -> /#/io.l5d.k8s.grpc/default/grpc/service/package\n    /svc => /$/io.buoyant.http.domainToPathPfx/srv ; # /svc/package.service -> /srv/service/package\n  client:\n    kind: io.l5d.static\n    configs:\n    # Always use TLS when sending to external grpc servers\n    - prefix: \"/$/inet/{service}\"\n      tls:\n        commonName: \"{service}\"\n\n- label: grpc-incoming\n  protocol: h2\n  servers:\n  - port: 4341\n    ip: 0.0.0.0\n  identifier:\n    kind: io.l5d.header.path\n    segments: 1\n  interpreter:\n    kind: default\n    transformers:\n    - kind: io.l5d.k8s.localnode\n      # hostNetwork: true # Uncomment if using host networking (eg for CNI)\n  dtab: |\n    /srv => /#/io.l5d.k8s/default/grpc ;             # /srv/service/package -> /#/io.l5d.k8s/default/grpc/service/package\n    /svc => /$/io.buoyant.http.domainToPathPfx/srv ; # /svc/package.service -> /srv/service/package\n\n# HTTP Ingress Controller listening on port 80\n- protocol: http\n  label: http-ingress\n  servers:\n    - port: 80\n      ip: 0.0.0.0\n      clearContext: true\n  identifier:\n    kind: io.l5d.ingress\n  dtab: /svc => /#/io.l5d.k8s\n\n# HTTP/2 Ingress Controller listening on port 8080\n- protocol: h2\n  label: h2-ingress\n  servers:\n    - port: 8080\n      ip: 0.0.0.0\n      clearContext: true\n  identifier:\n    kind: io.l5d.ingress\n  dtab: /svc => /#/io.l5d.k8s",
				},
			},
			&v1beta1.DaemonSet{
				TypeMeta: metav1.TypeMeta{
					Kind:       "DaemonSet",
					APIVersion: "extensions/v1beta1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "l5d",
					Namespace: "linkerd",
					Labels:    map[string]string{"app": "l5d"},
				},
				Spec: v1beta1.DaemonSetSpec{
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "l5d"},
						},
						Spec: v1.PodSpec{
							Volumes: []v1.Volume{
								{
									Name: "l5d-config",
									VolumeSource: v1.VolumeSource{
										ConfigMap: &v1.ConfigMapVolumeSource{
											LocalObjectReference: v1.LocalObjectReference{
												Name: "l5d-config",
											}},
									},
								},
							},
							Containers: []v1.Container{
								{
									Name:    "l5d",
									Image:   "buoyantio/linkerd:1.4.6",
									Command: nil,
									Args: []string{
										"/io.buoyant/linkerd/config/config.yaml",
									},
									WorkingDir: "",
									Ports: []v1.ContainerPort{
										{
											Name:          "http-outgoing",
											HostPort:      4140,
											ContainerPort: 4140,
											Protocol:      "",
											HostIP:        "",
										},
										{
											Name:          "http-incoming",
											HostPort:      0,
											ContainerPort: 4141,
											Protocol:      "",
											HostIP:        "",
										},
										{Name: "h2-outgoing", HostPort: 4240, ContainerPort: 4240},
										{Name: "h2-incoming", ContainerPort: 4241},
										{
											Name:          "grpc-outgoing",
											HostPort:      4340,
											ContainerPort: 4340,
										},
										{
											Name:          "grpc-incoming",
											ContainerPort: 4341,
										},
										{Name: "http-ingress", ContainerPort: 80},
										{Name: "h2-ingress", ContainerPort: 8080},
									},
									Env: []v1.EnvVar{
										{
											Name: "POD_IP",
											ValueFrom: &v1.EnvVarSource{
												FieldRef: &v1.ObjectFieldSelector{
													FieldPath: "status.podIP",
												},
											},
										},
										{
											Name: "NODE_NAME",
											ValueFrom: &v1.EnvVarSource{
												FieldRef: &v1.ObjectFieldSelector{
													FieldPath: "spec.nodeName",
												},
											},
										},
									},
									VolumeMounts: []v1.VolumeMount{
										{
											Name:      "l5d-config",
											ReadOnly:  true,
											MountPath: "/io.buoyant/linkerd/config",
										},
									},
								},
								{
									Name:  "kubectl",
									Image: "buoyantio/kubectl:v1.12.2",
									Args:  []string{"proxy", "-p", "8001"},
								},
							},
						},
					},
				},
			},
			&v1.Service{
				TypeMeta: metav1.TypeMeta{Kind: "Service", APIVersion: "v1"},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "l5d",
					Namespace: "linkerd",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{{
						Name: "http-outgoing",
						Port: 4140,
					},
						{
							Name: "http-incoming",
							Port: 4141,
						},
						{
							Name: "h2-outgoing",
							Port: 4240,
						},
						{
							Name: "h2-incoming",
							Port: 4241,
						},
						{
							Name: "grpc-outgoing",
							Port: 4340,
						},
						{
							Name: "grpc-incoming",
							Port: 4341,
						},
						{
							Name: "http-ingress",
							Port: 80,
						},
						{
							Name: "h2-ingress",
							Port: 8080,
						},
					},
					Selector: map[string]string{
						"app": "l5d",
					},
					Type: "LoadBalancer",
				},
			},
		}
		for i := range list {
			Expect(out[i]).To(Equal(list[i]))
		}
	})
})
