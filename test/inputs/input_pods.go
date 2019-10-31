package inputs

import (
	"github.com/gogo/protobuf/proto"
	"github.com/solo-io/solo-kit/pkg/api/external/kubernetes/pod"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BookInfoPodsIstioInject(svcNamespace string) kubernetes.PodList {
	pods := []kubev1.Pod{
		kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "details-v1-pod-1",
				Namespace: svcNamespace,
				Labels: map[string]string{
					"version": "v1",
					"app":     "details",
				},
			},
			Spec: kubev1.PodSpec{
				Volumes: []kubev1.Volume{
					kubev1.Volume{
						Name: "default-token-ztq95",
						VolumeSource: kubev1.VolumeSource{
							Secret: &kubev1.SecretVolumeSource{
								SecretName: "default-token-ztq95",
							},
						},
					},
					kubev1.Volume{
						Name: "istio-envoy",
						VolumeSource: kubev1.VolumeSource{
							EmptyDir: &kubev1.EmptyDirVolumeSource{
								Medium: "Memory",
							},
						},
					},
					kubev1.Volume{
						Name: "istio-certs",
						VolumeSource: kubev1.VolumeSource{
							Secret: &kubev1.SecretVolumeSource{
								SecretName: "istio.default",
							},
						},
					},
				},
				InitContainers: []kubev1.Container{
					kubev1.Container{
						Name:    "istio-init",
						Image:   "docker.io/istio/proxy_init:1.0.5",
						Command: []string{},
						Args: []string{
							"-p",
							"15001",
							"-u",
							"1337",
							"-m",
							"REDIRECT",
							"-i",
							"*",
							"-x",
							"",
							"-b",
							"9080",
							"-d",
							"",
						},
						WorkingDir: "",
						Ports:      []kubev1.ContainerPort{},
						EnvFrom:    []kubev1.EnvFromSource{},
						Env:        []kubev1.EnvVar{},
						Resources: kubev1.ResourceRequirements{
							Limits:   kubev1.ResourceList{},
							Requests: kubev1.ResourceList{},
						},
						VolumeMounts:             []kubev1.VolumeMount{},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						SecurityContext: &kubev1.SecurityContext{
							Capabilities: &kubev1.Capabilities{
								Add: []kubev1.Capability{
									"NET_ADMIN",
								},
								Drop: []kubev1.Capability{},
							},
							Privileged: proto.Bool(true),
						},
					},
				},
				Containers: []kubev1.Container{
					kubev1.Container{
						Name:       "details",
						Image:      "istio/examples-bookinfo-details-v1:1.8.0",
						Command:    []string{},
						Args:       []string{},
						WorkingDir: "",
						Ports: []kubev1.ContainerPort{
							kubev1.ContainerPort{
								Name:          "",
								HostPort:      0,
								ContainerPort: 9080,
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						EnvFrom: []kubev1.EnvFromSource{},
						Env:     []kubev1.EnvVar{},
						Resources: kubev1.ResourceRequirements{
							Limits:   kubev1.ResourceList{},
							Requests: kubev1.ResourceList{},
						},
						VolumeMounts: []kubev1.VolumeMount{
							kubev1.VolumeMount{
								Name:      "default-token-ztq95",
								ReadOnly:  true,
								MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								SubPath:   "",
							},
						},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						Stdin:                    false,
						StdinOnce:                false,
						TTY:                      false,
					},
					kubev1.Container{
						Name:    "istio-proxy",
						Image:   "docker.io/istio/proxyv2:1.0.5",
						Command: []string{},
						Args: []string{
							"proxy",
							"sidecar",
							"--configPath",
							"/etc/istio/proxy",
							"--binaryPath",
							"/usr/local/bin/envoy",
							"--serviceCluster",
							"details",
							"--drainDuration",
							"45s",
							"--parentShutdownDuration",
							"1m0s",
							"--discoveryAddress",
							"istio-pilot.istio-system:15007",
							"--discoveryRefreshDelay",
							"1s",
							"--zipkinAddress",
							"zipkin.istio-system:9411",
							"--connectTimeout",
							"10s",
							"--proxyAdminPort",
							"15000",
							"--controlPlaneAuthPolicy",
							"NONE",
						},
						WorkingDir: "",
						Ports: []kubev1.ContainerPort{
							kubev1.ContainerPort{
								Name:          "http-envoy-prom",
								HostPort:      0,
								ContainerPort: 15090,
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						EnvFrom: []kubev1.EnvFromSource{},
						Env: []kubev1.EnvVar{
							kubev1.EnvVar{
								Name:  "POD_NAME",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.name",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "POD_NAMESPACE",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.namespace",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "INSTANCE_IP",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "status.podIP",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "ISTIO_META_POD_NAME",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.name",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "ISTIO_META_INTERCEPTION_MODE",
								Value: "REDIRECT",
							},
							kubev1.EnvVar{
								Name:  "ISTIO_METAJSON_LABELS",
								Value: "{\"app\":\"details\",\"pod-template-hash\":\"2320667393\",\"version\":\"v1\"}\n",
							},
						},
						VolumeMounts: []kubev1.VolumeMount{
							kubev1.VolumeMount{
								Name:      "istio-envoy",
								ReadOnly:  false,
								MountPath: "/etc/istio/proxy",
								SubPath:   "",
							},
							kubev1.VolumeMount{
								Name:      "istio-certs",
								ReadOnly:  true,
								MountPath: "/etc/certs/",
								SubPath:   "",
							},
						},
					},
				},
				RestartPolicy: "Always",
			},
		},
		kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "petstore-pod-1",
				Namespace: svcNamespace,
				Labels: map[string]string{
					"app": "petstore",
				},
			},
			Spec: kubev1.PodSpec{
				Volumes: []kubev1.Volume{
					kubev1.Volume{
						Name: "default-token-ztq95",
						VolumeSource: kubev1.VolumeSource{
							Secret: &kubev1.SecretVolumeSource{
								SecretName: "default-token-ztq95",
							},
						},
					},
					kubev1.Volume{
						Name: "istio-envoy",
						VolumeSource: kubev1.VolumeSource{
							EmptyDir: &kubev1.EmptyDirVolumeSource{
								Medium: "Memory",
							},
						},
					},
					kubev1.Volume{
						Name: "istio-certs",
						VolumeSource: kubev1.VolumeSource{
							Secret: &kubev1.SecretVolumeSource{
								SecretName: "istio.default",
							},
						},
					},
				},
				InitContainers: []kubev1.Container{
					kubev1.Container{
						Name:    "istio-init",
						Image:   "docker.io/istio/proxy_init:1.0.5",
						Command: []string{},
						Args: []string{
							"-p",
							"15001",
							"-u",
							"1337",
							"-m",
							"REDIRECT",
							"-i",
							"*",
							"-x",
							"",
							"-b",
							"8080",
							"-d",
							"",
						},
						WorkingDir: "",
						Ports:      []kubev1.ContainerPort{},
						EnvFrom:    []kubev1.EnvFromSource{},
						Env:        []kubev1.EnvVar{},
						Resources: kubev1.ResourceRequirements{
							Limits:   kubev1.ResourceList{},
							Requests: kubev1.ResourceList{},
						},
						VolumeMounts:             []kubev1.VolumeMount{},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						SecurityContext: &kubev1.SecurityContext{
							Capabilities: &kubev1.Capabilities{
								Add: []kubev1.Capability{
									"NET_ADMIN",
								},
								Drop: []kubev1.Capability{},
							},
							Privileged: proto.Bool(true),
						},
						Stdin:     false,
						StdinOnce: false,
						TTY:       false,
					},
				},
				Containers: []kubev1.Container{
					kubev1.Container{
						Name:       "petstore",
						Image:      "soloio/petstore-example:latest",
						Command:    []string{},
						Args:       []string{},
						WorkingDir: "",
						Ports: []kubev1.ContainerPort{
							kubev1.ContainerPort{
								Name:          "http",
								HostPort:      0,
								ContainerPort: 8080,
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						EnvFrom: []kubev1.EnvFromSource{},
						Env:     []kubev1.EnvVar{},
						Resources: kubev1.ResourceRequirements{
							Limits:   kubev1.ResourceList{},
							Requests: kubev1.ResourceList{},
						},
						VolumeMounts: []kubev1.VolumeMount{
							kubev1.VolumeMount{
								Name:      "default-token-ztq95",
								ReadOnly:  true,
								MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								SubPath:   "",
							},
						},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "Always",
						Stdin:                    false,
						StdinOnce:                false,
						TTY:                      false,
					},
					kubev1.Container{
						Name:    "istio-proxy",
						Image:   "docker.io/istio/proxyv2:1.0.5",
						Command: []string{},
						Args: []string{
							"proxy",
							"sidecar",
							"--configPath",
							"/etc/istio/proxy",
							"--binaryPath",
							"/usr/local/bin/envoy",
							"--serviceCluster",
							"petstore",
							"--drainDuration",
							"45s",
							"--parentShutdownDuration",
							"1m0s",
							"--discoveryAddress",
							"istio-pilot.istio-system:15007",
							"--discoveryRefreshDelay",
							"1s",
							"--zipkinAddress",
							"zipkin.istio-system:9411",
							"--connectTimeout",
							"10s",
							"--proxyAdminPort",
							"15000",
							"--controlPlaneAuthPolicy",
							"NONE",
						},
						WorkingDir: "",
						Ports: []kubev1.ContainerPort{
							kubev1.ContainerPort{
								Name:          "http-envoy-prom",
								HostPort:      0,
								ContainerPort: 15090,
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						EnvFrom: []kubev1.EnvFromSource{},
						Env: []kubev1.EnvVar{
							kubev1.EnvVar{
								Name:  "POD_NAME",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.name",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "POD_NAMESPACE",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.namespace",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "INSTANCE_IP",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "status.podIP",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "ISTIO_META_POD_NAME",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.name",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "ISTIO_META_INTERCEPTION_MODE",
								Value: "REDIRECT",
							},
							kubev1.EnvVar{
								Name:  "ISTIO_METAJSON_LABELS",
								Value: "{\"app\":\"petstore\",\"pod-template-hash\":\"29840675\"}\n",
							},
						},
						VolumeMounts: []kubev1.VolumeMount{
							kubev1.VolumeMount{
								Name:      "istio-envoy",
								ReadOnly:  false,
								MountPath: "/etc/istio/proxy",
								SubPath:   "",
							},
							kubev1.VolumeMount{
								Name:      "istio-certs",
								ReadOnly:  true,
								MountPath: "/etc/certs/",
								SubPath:   "",
							},
						},
					},
				},
			},
		},
		kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "productpage-v1-pod-1",
				Namespace: svcNamespace,
				Labels: map[string]string{
					"pod-template-hash": "106465911",
					"version":           "v1",
					"app":               "productpage",
				},
			},
			Spec: kubev1.PodSpec{
				Volumes: []kubev1.Volume{
					kubev1.Volume{
						Name: "default-token-ztq95",
						VolumeSource: kubev1.VolumeSource{
							Secret: &kubev1.SecretVolumeSource{
								SecretName: "default-token-ztq95",
							},
						},
					},
					kubev1.Volume{
						Name: "istio-envoy",
						VolumeSource: kubev1.VolumeSource{
							EmptyDir: &kubev1.EmptyDirVolumeSource{
								Medium: "Memory",
							},
						},
					},
					kubev1.Volume{
						Name: "istio-certs",
						VolumeSource: kubev1.VolumeSource{
							Secret: &kubev1.SecretVolumeSource{
								SecretName: "istio.default",
							},
						},
					},
				},
				InitContainers: []kubev1.Container{
					kubev1.Container{
						Name:    "istio-init",
						Image:   "docker.io/istio/proxy_init:1.0.5",
						Command: []string{},
						Args: []string{
							"-p",
							"15001",
							"-u",
							"1337",
							"-m",
							"REDIRECT",
							"-i",
							"*",
							"-x",
							"",
							"-b",
							"9080",
							"-d",
							"",
						},
						WorkingDir: "",
						Ports:      []kubev1.ContainerPort{},
						EnvFrom:    []kubev1.EnvFromSource{},
						Env:        []kubev1.EnvVar{},
						Resources: kubev1.ResourceRequirements{
							Limits:   kubev1.ResourceList{},
							Requests: kubev1.ResourceList{},
						},
						VolumeMounts:             []kubev1.VolumeMount{},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						SecurityContext: &kubev1.SecurityContext{
							Capabilities: &kubev1.Capabilities{
								Add: []kubev1.Capability{
									"NET_ADMIN",
								},
								Drop: []kubev1.Capability{},
							},
							Privileged: proto.Bool(true),
						},
						Stdin:     false,
						StdinOnce: false,
						TTY:       false,
					},
				},
				Containers: []kubev1.Container{
					kubev1.Container{
						Name:       "productpage",
						Image:      "istio/examples-bookinfo-productpage-v1:1.8.0",
						Command:    []string{},
						Args:       []string{},
						WorkingDir: "",
						Ports: []kubev1.ContainerPort{
							kubev1.ContainerPort{
								Name:          "",
								HostPort:      0,
								ContainerPort: 9080,
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						EnvFrom: []kubev1.EnvFromSource{},
						Env:     []kubev1.EnvVar{},
						Resources: kubev1.ResourceRequirements{
							Limits:   kubev1.ResourceList{},
							Requests: kubev1.ResourceList{},
						},
						VolumeMounts: []kubev1.VolumeMount{
							kubev1.VolumeMount{
								Name:      "default-token-ztq95",
								ReadOnly:  true,
								MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								SubPath:   "",
							},
						},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						Stdin:                    false,
						StdinOnce:                false,
						TTY:                      false,
					},
					kubev1.Container{
						Name:    "istio-proxy",
						Image:   "docker.io/istio/proxyv2:1.0.5",
						Command: []string{},
						Args: []string{
							"proxy",
							"sidecar",
							"--configPath",
							"/etc/istio/proxy",
							"--binaryPath",
							"/usr/local/bin/envoy",
							"--serviceCluster",
							"productpage",
							"--drainDuration",
							"45s",
							"--parentShutdownDuration",
							"1m0s",
							"--discoveryAddress",
							"istio-pilot.istio-system:15007",
							"--discoveryRefreshDelay",
							"1s",
							"--zipkinAddress",
							"zipkin.istio-system:9411",
							"--connectTimeout",
							"10s",
							"--proxyAdminPort",
							"15000",
							"--controlPlaneAuthPolicy",
							"NONE",
						},
						WorkingDir: "",
						Ports: []kubev1.ContainerPort{
							kubev1.ContainerPort{
								Name:          "http-envoy-prom",
								HostPort:      0,
								ContainerPort: 15090,
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						EnvFrom: []kubev1.EnvFromSource{},
						Env: []kubev1.EnvVar{
							kubev1.EnvVar{
								Name:  "POD_NAME",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.name",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "POD_NAMESPACE",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.namespace",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "INSTANCE_IP",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "status.podIP",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "ISTIO_META_POD_NAME",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.name",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "ISTIO_META_INTERCEPTION_MODE",
								Value: "REDIRECT",
							},
							kubev1.EnvVar{
								Name:  "ISTIO_METAJSON_LABELS",
								Value: "{\"app\":\"productpage\",\"pod-template-hash\":\"106465911\",\"version\":\"v1\"}\n",
							},
						},
					},
				},
				RestartPolicy: "Always",
			},
		},
		kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ratings-v1-pod-1",
				Namespace: svcNamespace,
				Labels: map[string]string{
					"version":           "v1",
					"app":               "ratings",
					"pod-template-hash": "36741505",
				},
			},
			Spec: kubev1.PodSpec{
				Volumes: []kubev1.Volume{
					kubev1.Volume{
						Name: "default-token-ztq95",
						VolumeSource: kubev1.VolumeSource{
							Secret: &kubev1.SecretVolumeSource{
								SecretName: "default-token-ztq95",
							},
						},
					},
					kubev1.Volume{
						Name: "istio-envoy",
						VolumeSource: kubev1.VolumeSource{
							EmptyDir: &kubev1.EmptyDirVolumeSource{
								Medium: "Memory",
							},
						},
					},
					kubev1.Volume{
						Name: "istio-certs",
						VolumeSource: kubev1.VolumeSource{
							Secret: &kubev1.SecretVolumeSource{
								SecretName: "istio.default",
							},
						},
					},
				},
				InitContainers: []kubev1.Container{
					kubev1.Container{
						Name:    "istio-init",
						Image:   "docker.io/istio/proxy_init:1.0.5",
						Command: []string{},
						Args: []string{
							"-p",
							"15001",
							"-u",
							"1337",
							"-m",
							"REDIRECT",
							"-i",
							"*",
							"-x",
							"",
							"-b",
							"9080",
							"-d",
							"",
						},
						WorkingDir: "",
						Ports:      []kubev1.ContainerPort{},
						EnvFrom:    []kubev1.EnvFromSource{},
						Env:        []kubev1.EnvVar{},
						Resources: kubev1.ResourceRequirements{
							Limits:   kubev1.ResourceList{},
							Requests: kubev1.ResourceList{},
						},
						VolumeMounts:             []kubev1.VolumeMount{},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						SecurityContext: &kubev1.SecurityContext{
							Capabilities: &kubev1.Capabilities{
								Add: []kubev1.Capability{
									"NET_ADMIN",
								},
								Drop: []kubev1.Capability{},
							},
							Privileged: proto.Bool(true),
						},
						Stdin:     false,
						StdinOnce: false,
						TTY:       false,
					},
				},
				Containers: []kubev1.Container{
					kubev1.Container{
						Name:       "ratings",
						Image:      "istio/examples-bookinfo-ratings-v1:1.8.0",
						Command:    []string{},
						Args:       []string{},
						WorkingDir: "",
						Ports: []kubev1.ContainerPort{
							kubev1.ContainerPort{
								Name:          "",
								HostPort:      0,
								ContainerPort: 9080,
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						EnvFrom: []kubev1.EnvFromSource{},
						Env:     []kubev1.EnvVar{},
						Resources: kubev1.ResourceRequirements{
							Limits:   kubev1.ResourceList{},
							Requests: kubev1.ResourceList{},
						},
						VolumeMounts: []kubev1.VolumeMount{
							kubev1.VolumeMount{
								Name:      "default-token-ztq95",
								ReadOnly:  true,
								MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								SubPath:   "",
							},
						},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						Stdin:                    false,
						StdinOnce:                false,
						TTY:                      false,
					},
					kubev1.Container{
						Name:    "istio-proxy",
						Image:   "docker.io/istio/proxyv2:1.0.5",
						Command: []string{},
						Args: []string{
							"proxy",
							"sidecar",
							"--configPath",
							"/etc/istio/proxy",
							"--binaryPath",
							"/usr/local/bin/envoy",
							"--serviceCluster",
							"ratings",
							"--drainDuration",
							"45s",
							"--parentShutdownDuration",
							"1m0s",
							"--discoveryAddress",
							"istio-pilot.istio-system:15007",
							"--discoveryRefreshDelay",
							"1s",
							"--zipkinAddress",
							"zipkin.istio-system:9411",
							"--connectTimeout",
							"10s",
							"--proxyAdminPort",
							"15000",
							"--controlPlaneAuthPolicy",
							"NONE",
						},
						WorkingDir: "",
						Ports: []kubev1.ContainerPort{
							kubev1.ContainerPort{
								Name:          "http-envoy-prom",
								HostPort:      0,
								ContainerPort: 15090,
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						EnvFrom: []kubev1.EnvFromSource{},
						Env: []kubev1.EnvVar{
							kubev1.EnvVar{
								Name:  "POD_NAME",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.name",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "POD_NAMESPACE",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.namespace",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "INSTANCE_IP",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "status.podIP",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "ISTIO_META_POD_NAME",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.name",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "ISTIO_META_INTERCEPTION_MODE",
								Value: "REDIRECT",
							},
							kubev1.EnvVar{
								Name:  "ISTIO_METAJSON_LABELS",
								Value: "{\"app\":\"ratings\",\"pod-template-hash\":\"36741505\",\"version\":\"v1\"}\n",
							},
						},
					},
				},
				RestartPolicy: "Always",
				DNSPolicy:     "ClusterFirst",
			},
		},
		kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "reviews-v1-pod-1",
				Namespace: svcNamespace,
				Labels: map[string]string{
					"app":               "reviews",
					"pod-template-hash": "986923066",
					"version":           "v1",
				},
			},
			Spec: kubev1.PodSpec{
				Volumes: []kubev1.Volume{
					kubev1.Volume{
						Name: "default-token-ztq95",
						VolumeSource: kubev1.VolumeSource{
							Secret: &kubev1.SecretVolumeSource{
								SecretName: "default-token-ztq95",
							},
						},
					},
					kubev1.Volume{
						Name: "istio-envoy",
						VolumeSource: kubev1.VolumeSource{
							EmptyDir: &kubev1.EmptyDirVolumeSource{
								Medium: "Memory",
							},
						},
					},
					kubev1.Volume{
						Name: "istio-certs",
						VolumeSource: kubev1.VolumeSource{
							Secret: &kubev1.SecretVolumeSource{
								SecretName: "istio.default",
							},
						},
					},
				},
				InitContainers: []kubev1.Container{
					kubev1.Container{
						Name:    "istio-init",
						Image:   "docker.io/istio/proxy_init:1.0.5",
						Command: []string{},
						Args: []string{
							"-p",
							"15001",
							"-u",
							"1337",
							"-m",
							"REDIRECT",
							"-i",
							"*",
							"-x",
							"",
							"-b",
							"9080",
							"-d",
							"",
						},
						WorkingDir: "",
						Ports:      []kubev1.ContainerPort{},
						EnvFrom:    []kubev1.EnvFromSource{},
						Env:        []kubev1.EnvVar{},
						Resources: kubev1.ResourceRequirements{
							Limits:   kubev1.ResourceList{},
							Requests: kubev1.ResourceList{},
						},
						VolumeMounts:             []kubev1.VolumeMount{},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						SecurityContext: &kubev1.SecurityContext{
							Capabilities: &kubev1.Capabilities{
								Add: []kubev1.Capability{
									"NET_ADMIN",
								},
								Drop: []kubev1.Capability{},
							},
							Privileged: proto.Bool(true),
						},
						Stdin:     false,
						StdinOnce: false,
						TTY:       false,
					},
				},
				Containers: []kubev1.Container{
					kubev1.Container{
						Name:       "reviews",
						Image:      "istio/examples-bookinfo-reviews-v1:1.8.0",
						Command:    []string{},
						Args:       []string{},
						WorkingDir: "",
						Ports: []kubev1.ContainerPort{
							kubev1.ContainerPort{
								Name:          "",
								HostPort:      0,
								ContainerPort: 9080,
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						EnvFrom: []kubev1.EnvFromSource{},
						Env:     []kubev1.EnvVar{},
						Resources: kubev1.ResourceRequirements{
							Limits:   kubev1.ResourceList{},
							Requests: kubev1.ResourceList{},
						},
						VolumeMounts: []kubev1.VolumeMount{
							kubev1.VolumeMount{
								Name:      "default-token-ztq95",
								ReadOnly:  true,
								MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								SubPath:   "",
							},
						},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						Stdin:                    false,
						StdinOnce:                false,
						TTY:                      false,
					},
					kubev1.Container{
						Name:    "istio-proxy",
						Image:   "docker.io/istio/proxyv2:1.0.5",
						Command: []string{},
						Args: []string{
							"proxy",
							"sidecar",
							"--configPath",
							"/etc/istio/proxy",
							"--binaryPath",
							"/usr/local/bin/envoy",
							"--serviceCluster",
							"reviews",
							"--drainDuration",
							"45s",
							"--parentShutdownDuration",
							"1m0s",
							"--discoveryAddress",
							"istio-pilot.istio-system:15007",
							"--discoveryRefreshDelay",
							"1s",
							"--zipkinAddress",
							"zipkin.istio-system:9411",
							"--connectTimeout",
							"10s",
							"--proxyAdminPort",
							"15000",
							"--controlPlaneAuthPolicy",
							"NONE",
						},
						WorkingDir: "",
						Ports: []kubev1.ContainerPort{
							kubev1.ContainerPort{
								Name:          "http-envoy-prom",
								HostPort:      0,
								ContainerPort: 15090,
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						EnvFrom: []kubev1.EnvFromSource{},
						Env: []kubev1.EnvVar{
							kubev1.EnvVar{
								Name:  "POD_NAME",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.name",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "POD_NAMESPACE",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.namespace",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "INSTANCE_IP",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "status.podIP",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "ISTIO_META_POD_NAME",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.name",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "ISTIO_META_INTERCEPTION_MODE",
								Value: "REDIRECT",
							},
							kubev1.EnvVar{
								Name:  "ISTIO_METAJSON_LABELS",
								Value: "{\"app\":\"reviews\",\"pod-template-hash\":\"986923066\",\"version\":\"v1\"}\n",
							},
						},
						VolumeMounts: []kubev1.VolumeMount{
							kubev1.VolumeMount{
								Name:      "istio-envoy",
								ReadOnly:  false,
								MountPath: "/etc/istio/proxy",
								SubPath:   "",
							},
							kubev1.VolumeMount{
								Name:      "istio-certs",
								ReadOnly:  true,
								MountPath: "/etc/certs/",
								SubPath:   "",
							},
						},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						Stdin:                    false,
						StdinOnce:                false,
						TTY:                      false,
					},
				},
				RestartPolicy: "Always",
			},
		},
		kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "reviews-v2-pod-1",
				Namespace: svcNamespace,
				Labels: map[string]string{
					"app":               "reviews",
					"pod-template-hash": "1687143382",
					"version":           "v2",
				},
			},
			Spec: kubev1.PodSpec{
				Volumes: []kubev1.Volume{
					kubev1.Volume{
						Name: "default-token-ztq95",
						VolumeSource: kubev1.VolumeSource{
							Secret: &kubev1.SecretVolumeSource{
								SecretName: "default-token-ztq95",
							},
						},
					},
					kubev1.Volume{
						Name: "istio-envoy",
						VolumeSource: kubev1.VolumeSource{
							EmptyDir: &kubev1.EmptyDirVolumeSource{
								Medium: "Memory",
							},
						},
					},
					kubev1.Volume{
						Name: "istio-certs",
						VolumeSource: kubev1.VolumeSource{
							Secret: &kubev1.SecretVolumeSource{
								SecretName: "istio.default",
							},
						},
					},
				},
				InitContainers: []kubev1.Container{
					kubev1.Container{
						Name:    "istio-init",
						Image:   "docker.io/istio/proxy_init:1.0.5",
						Command: []string{},
						Args: []string{
							"-p",
							"15001",
							"-u",
							"1337",
							"-m",
							"REDIRECT",
							"-i",
							"*",
							"-x",
							"",
							"-b",
							"9080",
							"-d",
							"",
						},
						WorkingDir: "",
						Ports:      []kubev1.ContainerPort{},
						EnvFrom:    []kubev1.EnvFromSource{},
						Env:        []kubev1.EnvVar{},
						Resources: kubev1.ResourceRequirements{
							Limits:   kubev1.ResourceList{},
							Requests: kubev1.ResourceList{},
						},
						VolumeMounts:             []kubev1.VolumeMount{},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						SecurityContext: &kubev1.SecurityContext{
							Capabilities: &kubev1.Capabilities{
								Add: []kubev1.Capability{
									"NET_ADMIN",
								},
								Drop: []kubev1.Capability{},
							},
							Privileged: proto.Bool(true),
						},
						Stdin:     false,
						StdinOnce: false,
						TTY:       false,
					},
				},
				Containers: []kubev1.Container{
					kubev1.Container{
						Name:       "reviews",
						Image:      "istio/examples-bookinfo-reviews-v2:1.8.0",
						Command:    []string{},
						Args:       []string{},
						WorkingDir: "",
						Ports: []kubev1.ContainerPort{
							kubev1.ContainerPort{
								Name:          "",
								HostPort:      0,
								ContainerPort: 9080,
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						EnvFrom: []kubev1.EnvFromSource{},
						Env:     []kubev1.EnvVar{},
						Resources: kubev1.ResourceRequirements{
							Limits:   kubev1.ResourceList{},
							Requests: kubev1.ResourceList{},
						},
						VolumeMounts: []kubev1.VolumeMount{
							kubev1.VolumeMount{
								Name:      "default-token-ztq95",
								ReadOnly:  true,
								MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								SubPath:   "",
							},
						},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						Stdin:                    false,
						StdinOnce:                false,
						TTY:                      false,
					},
					kubev1.Container{
						Name:    "istio-proxy",
						Image:   "docker.io/istio/proxyv2:1.0.5",
						Command: []string{},
						Args: []string{
							"proxy",
							"sidecar",
							"--configPath",
							"/etc/istio/proxy",
							"--binaryPath",
							"/usr/local/bin/envoy",
							"--serviceCluster",
							"reviews",
							"--drainDuration",
							"45s",
							"--parentShutdownDuration",
							"1m0s",
							"--discoveryAddress",
							"istio-pilot.istio-system:15007",
							"--discoveryRefreshDelay",
							"1s",
							"--zipkinAddress",
							"zipkin.istio-system:9411",
							"--connectTimeout",
							"10s",
							"--proxyAdminPort",
							"15000",
							"--controlPlaneAuthPolicy",
							"NONE",
						},
						WorkingDir: "",
						Ports: []kubev1.ContainerPort{
							kubev1.ContainerPort{
								Name:          "http-envoy-prom",
								HostPort:      0,
								ContainerPort: 15090,
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						EnvFrom: []kubev1.EnvFromSource{},
						Env: []kubev1.EnvVar{
							kubev1.EnvVar{
								Name:  "POD_NAME",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.name",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "POD_NAMESPACE",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.namespace",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "INSTANCE_IP",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "status.podIP",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "ISTIO_META_POD_NAME",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.name",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "ISTIO_META_INTERCEPTION_MODE",
								Value: "REDIRECT",
							},
							kubev1.EnvVar{
								Name:  "ISTIO_METAJSON_LABELS",
								Value: "{\"app\":\"reviews\",\"pod-template-hash\":\"1687143382\",\"version\":\"v2\"}\n",
							},
						},
						VolumeMounts: []kubev1.VolumeMount{
							kubev1.VolumeMount{
								Name:      "istio-envoy",
								ReadOnly:  false,
								MountPath: "/etc/istio/proxy",
								SubPath:   "",
							},
							kubev1.VolumeMount{
								Name:      "istio-certs",
								ReadOnly:  true,
								MountPath: "/etc/certs/",
								SubPath:   "",
							},
						},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						Stdin:                    false,
						StdinOnce:                false,
						TTY:                      false,
					},
				},
				RestartPolicy: "Always",
				DNSPolicy:     "ClusterFirst",
				NodeSelector:  map[string]string{},
				NodeName:      "minikube",
			},
		},
		kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "reviews-v3-pod-1",
				Namespace: svcNamespace,
				Labels: map[string]string{
					"app":               "reviews",
					"pod-template-hash": "884027734",
					"version":           "v3",
				},
			},
			Spec: kubev1.PodSpec{
				Volumes: []kubev1.Volume{
					kubev1.Volume{
						Name: "default-token-ztq95",
						VolumeSource: kubev1.VolumeSource{
							Secret: &kubev1.SecretVolumeSource{
								SecretName: "default-token-ztq95",
							},
						},
					},
					kubev1.Volume{
						Name: "istio-envoy",
						VolumeSource: kubev1.VolumeSource{
							EmptyDir: &kubev1.EmptyDirVolumeSource{
								Medium: "Memory",
							},
						},
					},
					kubev1.Volume{
						Name: "istio-certs",
						VolumeSource: kubev1.VolumeSource{
							Secret: &kubev1.SecretVolumeSource{
								SecretName: "istio.default",
							},
						},
					},
				},
				InitContainers: []kubev1.Container{
					kubev1.Container{
						Name:    "istio-init",
						Image:   "docker.io/istio/proxy_init:1.0.5",
						Command: []string{},
						Args: []string{
							"-p",
							"15001",
							"-u",
							"1337",
							"-m",
							"REDIRECT",
							"-i",
							"*",
							"-x",
							"",
							"-b",
							"9080",
							"-d",
							"",
						},
						WorkingDir: "",
						Ports:      []kubev1.ContainerPort{},
						EnvFrom:    []kubev1.EnvFromSource{},
						Env:        []kubev1.EnvVar{},
						Resources: kubev1.ResourceRequirements{
							Limits:   kubev1.ResourceList{},
							Requests: kubev1.ResourceList{},
						},
						VolumeMounts:             []kubev1.VolumeMount{},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						SecurityContext: &kubev1.SecurityContext{
							Capabilities: &kubev1.Capabilities{
								Add: []kubev1.Capability{
									"NET_ADMIN",
								},
								Drop: []kubev1.Capability{},
							},
							Privileged: proto.Bool(true),
						},
						Stdin:     false,
						StdinOnce: false,
						TTY:       false,
					},
				},
				Containers: []kubev1.Container{
					kubev1.Container{
						Name:       "reviews",
						Image:      "istio/examples-bookinfo-reviews-v3:1.8.0",
						Command:    []string{},
						Args:       []string{},
						WorkingDir: "",
						Ports: []kubev1.ContainerPort{
							kubev1.ContainerPort{
								Name:          "",
								HostPort:      0,
								ContainerPort: 9080,
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						EnvFrom: []kubev1.EnvFromSource{},
						Env:     []kubev1.EnvVar{},
						Resources: kubev1.ResourceRequirements{
							Limits:   kubev1.ResourceList{},
							Requests: kubev1.ResourceList{},
						},
						VolumeMounts: []kubev1.VolumeMount{
							kubev1.VolumeMount{
								Name:      "default-token-ztq95",
								ReadOnly:  true,
								MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								SubPath:   "",
							},
						},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						Stdin:                    false,
						StdinOnce:                false,
						TTY:                      false,
					},
					kubev1.Container{
						Name:    "istio-proxy",
						Image:   "docker.io/istio/proxyv2:1.0.5",
						Command: []string{},
						Args: []string{
							"proxy",
							"sidecar",
							"--configPath",
							"/etc/istio/proxy",
							"--binaryPath",
							"/usr/local/bin/envoy",
							"--serviceCluster",
							"reviews",
							"--drainDuration",
							"45s",
							"--parentShutdownDuration",
							"1m0s",
							"--discoveryAddress",
							"istio-pilot.istio-system:15007",
							"--discoveryRefreshDelay",
							"1s",
							"--zipkinAddress",
							"zipkin.istio-system:9411",
							"--connectTimeout",
							"10s",
							"--proxyAdminPort",
							"15000",
							"--controlPlaneAuthPolicy",
							"NONE",
						},
						WorkingDir: "",
						Ports: []kubev1.ContainerPort{
							kubev1.ContainerPort{
								Name:          "http-envoy-prom",
								HostPort:      0,
								ContainerPort: 15090,
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						EnvFrom: []kubev1.EnvFromSource{},
						Env: []kubev1.EnvVar{
							kubev1.EnvVar{
								Name:  "POD_NAME",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.name",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "POD_NAMESPACE",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.namespace",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "INSTANCE_IP",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "status.podIP",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "ISTIO_META_POD_NAME",
								Value: "",
								ValueFrom: &kubev1.EnvVarSource{
									FieldRef: &kubev1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.name",
									},
								},
							},
							kubev1.EnvVar{
								Name:  "ISTIO_META_INTERCEPTION_MODE",
								Value: "REDIRECT",
							},
							kubev1.EnvVar{
								Name:  "ISTIO_METAJSON_LABELS",
								Value: "{\"app\":\"reviews\",\"pod-template-hash\":\"884027734\",\"version\":\"v3\"}\n",
							},
						},
						VolumeMounts: []kubev1.VolumeMount{
							kubev1.VolumeMount{
								Name:      "istio-envoy",
								ReadOnly:  false,
								MountPath: "/etc/istio/proxy",
								SubPath:   "",
							},
							kubev1.VolumeMount{
								Name:      "istio-certs",
								ReadOnly:  true,
								MountPath: "/etc/certs/",
								SubPath:   "",
							},
						},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						Stdin:                    false,
						StdinOnce:                false,
						TTY:                      false,
					},
				},
				RestartPolicy: "Always",
				DNSPolicy:     "ClusterFirst",
				NodeSelector:  map[string]string{},
			},
		},
	}
	var customPods kubernetes.PodList
	for _, p := range pods {
		// set service account
		p.Spec.ServiceAccountName = p.Name

		convertedPod := pod.FromKubePod(&p)

		customPods = append(customPods, convertedPod)
	}
	return customPods
}

func BookInfoPodsLinkerdInject(svcNamespace string) kubernetes.PodList {
	pods := []kubev1.Pod{
		kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "details-v1-pod-1",
				Namespace: svcNamespace,
				Labels: map[string]string{
					"version": "v1",
					"app":     "details",
				},
			},
			Spec: kubev1.PodSpec{
				Containers: []kubev1.Container{
					kubev1.Container{
						Name:       "details",
						Image:      "istio/examples-bookinfo-details-v1:1.8.0",
						Command:    []string{},
						Args:       []string{},
						WorkingDir: "",
						Ports: []kubev1.ContainerPort{
							kubev1.ContainerPort{
								Name:          "",
								HostPort:      0,
								ContainerPort: 9080,
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						EnvFrom: []kubev1.EnvFromSource{},
						Env:     []kubev1.EnvVar{},
						Resources: kubev1.ResourceRequirements{
							Limits:   kubev1.ResourceList{},
							Requests: kubev1.ResourceList{},
						},
						VolumeMounts: []kubev1.VolumeMount{
							kubev1.VolumeMount{
								Name:      "default-token-ztq95",
								ReadOnly:  true,
								MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								SubPath:   "",
							},
						},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						Stdin:                    false,
						StdinOnce:                false,
						TTY:                      false,
					},
				},
				RestartPolicy: "Always",
			},
		},
		kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "petstore-pod-1",
				Namespace: svcNamespace,
				Labels: map[string]string{
					"app": "petstore",
				},
			},
			Spec: kubev1.PodSpec{
				Containers: []kubev1.Container{
					kubev1.Container{
						Name:       "petstore",
						Image:      "soloio/petstore-example:latest",
						Command:    []string{},
						Args:       []string{},
						WorkingDir: "",
						Ports: []kubev1.ContainerPort{
							kubev1.ContainerPort{
								Name:          "http",
								HostPort:      0,
								ContainerPort: 8080,
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						EnvFrom: []kubev1.EnvFromSource{},
						Env:     []kubev1.EnvVar{},
						Resources: kubev1.ResourceRequirements{
							Limits:   kubev1.ResourceList{},
							Requests: kubev1.ResourceList{},
						},
						VolumeMounts: []kubev1.VolumeMount{
							kubev1.VolumeMount{
								Name:      "default-token-ztq95",
								ReadOnly:  true,
								MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								SubPath:   "",
							},
						},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "Always",
						Stdin:                    false,
						StdinOnce:                false,
						TTY:                      false,
					},
				},
			},
		},
		kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "productpage-v1-pod-1",
				Namespace: svcNamespace,
				Labels: map[string]string{
					"pod-template-hash": "106465911",
					"version":           "v1",
					"app":               "productpage",
				},
			},
			Spec: kubev1.PodSpec{
				Containers: []kubev1.Container{
					kubev1.Container{
						Name:       "productpage",
						Image:      "istio/examples-bookinfo-productpage-v1:1.8.0",
						Command:    []string{},
						Args:       []string{},
						WorkingDir: "",
						Ports: []kubev1.ContainerPort{
							kubev1.ContainerPort{
								Name:          "",
								HostPort:      0,
								ContainerPort: 9080,
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						EnvFrom: []kubev1.EnvFromSource{},
						Env:     []kubev1.EnvVar{},
						Resources: kubev1.ResourceRequirements{
							Limits:   kubev1.ResourceList{},
							Requests: kubev1.ResourceList{},
						},
						VolumeMounts: []kubev1.VolumeMount{
							kubev1.VolumeMount{
								Name:      "default-token-ztq95",
								ReadOnly:  true,
								MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								SubPath:   "",
							},
						},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						Stdin:                    false,
						StdinOnce:                false,
						TTY:                      false,
					},
				},
				RestartPolicy: "Always",
			},
		},
		kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ratings-v1-pod-1",
				Namespace: svcNamespace,
				Labels: map[string]string{
					"version":           "v1",
					"app":               "ratings",
					"pod-template-hash": "36741505",
				},
			},
			Spec: kubev1.PodSpec{
				Containers: []kubev1.Container{
					kubev1.Container{
						Name:       "ratings",
						Image:      "istio/examples-bookinfo-ratings-v1:1.8.0",
						Command:    []string{},
						Args:       []string{},
						WorkingDir: "",
						Ports: []kubev1.ContainerPort{
							kubev1.ContainerPort{
								Name:          "",
								HostPort:      0,
								ContainerPort: 9080,
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						EnvFrom: []kubev1.EnvFromSource{},
						Env:     []kubev1.EnvVar{},
						Resources: kubev1.ResourceRequirements{
							Limits:   kubev1.ResourceList{},
							Requests: kubev1.ResourceList{},
						},
						VolumeMounts: []kubev1.VolumeMount{
							kubev1.VolumeMount{
								Name:      "default-token-ztq95",
								ReadOnly:  true,
								MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								SubPath:   "",
							},
						},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						Stdin:                    false,
						StdinOnce:                false,
						TTY:                      false,
					},
				},
				RestartPolicy: "Always",
				DNSPolicy:     "ClusterFirst",
			},
		},
		kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "reviews-v1-pod-1",
				Namespace: svcNamespace,
				Labels: map[string]string{
					"app":               "reviews",
					"pod-template-hash": "986923066",
					"version":           "v1",
				},
			},
			Spec: kubev1.PodSpec{
				Containers: []kubev1.Container{
					kubev1.Container{
						Name:       "reviews",
						Image:      "istio/examples-bookinfo-reviews-v1:1.8.0",
						Command:    []string{},
						Args:       []string{},
						WorkingDir: "",
						Ports: []kubev1.ContainerPort{
							kubev1.ContainerPort{
								Name:          "",
								HostPort:      0,
								ContainerPort: 9080,
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						EnvFrom: []kubev1.EnvFromSource{},
						Env:     []kubev1.EnvVar{},
						Resources: kubev1.ResourceRequirements{
							Limits:   kubev1.ResourceList{},
							Requests: kubev1.ResourceList{},
						},
						VolumeMounts: []kubev1.VolumeMount{
							kubev1.VolumeMount{
								Name:      "default-token-ztq95",
								ReadOnly:  true,
								MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								SubPath:   "",
							},
						},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						Stdin:                    false,
						StdinOnce:                false,
						TTY:                      false,
					},
				},
				RestartPolicy: "Always",
			},
		},
		kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "reviews-v2-pod-1",
				Namespace: svcNamespace,
				Labels: map[string]string{
					"app":               "reviews",
					"pod-template-hash": "1687143382",
					"version":           "v2",
				},
			},
			Spec: kubev1.PodSpec{
				Volumes: []kubev1.Volume{
					kubev1.Volume{
						Name: "default-token-ztq95",
						VolumeSource: kubev1.VolumeSource{
							Secret: &kubev1.SecretVolumeSource{
								SecretName: "default-token-ztq95",
							},
						},
					},
					kubev1.Volume{
						Name: "istio-envoy",
						VolumeSource: kubev1.VolumeSource{
							EmptyDir: &kubev1.EmptyDirVolumeSource{
								Medium: "Memory",
							},
						},
					},
					kubev1.Volume{
						Name: "istio-certs",
						VolumeSource: kubev1.VolumeSource{
							Secret: &kubev1.SecretVolumeSource{
								SecretName: "istio.default",
							},
						},
					},
				},
				Containers: []kubev1.Container{
					kubev1.Container{
						Name:       "reviews",
						Image:      "istio/examples-bookinfo-reviews-v2:1.8.0",
						Command:    []string{},
						Args:       []string{},
						WorkingDir: "",
						Ports: []kubev1.ContainerPort{
							kubev1.ContainerPort{
								Name:          "",
								HostPort:      0,
								ContainerPort: 9080,
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						EnvFrom: []kubev1.EnvFromSource{},
						Env:     []kubev1.EnvVar{},
						Resources: kubev1.ResourceRequirements{
							Limits:   kubev1.ResourceList{},
							Requests: kubev1.ResourceList{},
						},
						VolumeMounts: []kubev1.VolumeMount{
							kubev1.VolumeMount{
								Name:      "default-token-ztq95",
								ReadOnly:  true,
								MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								SubPath:   "",
							},
						},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						Stdin:                    false,
						StdinOnce:                false,
						TTY:                      false,
					},
				},
				RestartPolicy: "Always",
				DNSPolicy:     "ClusterFirst",
				NodeSelector:  map[string]string{},
				NodeName:      "minikube",
			},
		},
		kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "reviews-v3-pod-1",
				Namespace: svcNamespace,
				Labels: map[string]string{
					"app":               "reviews",
					"pod-template-hash": "884027734",
					"version":           "v3",
				},
			},
			Spec: kubev1.PodSpec{
				Containers: []kubev1.Container{
					kubev1.Container{
						Name:       "reviews",
						Image:      "istio/examples-bookinfo-reviews-v3:1.8.0",
						Command:    []string{},
						Args:       []string{},
						WorkingDir: "",
						Ports: []kubev1.ContainerPort{
							kubev1.ContainerPort{
								Name:          "",
								HostPort:      0,
								ContainerPort: 9080,
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						EnvFrom: []kubev1.EnvFromSource{},
						Env:     []kubev1.EnvVar{},
						Resources: kubev1.ResourceRequirements{
							Limits:   kubev1.ResourceList{},
							Requests: kubev1.ResourceList{},
						},
						VolumeMounts: []kubev1.VolumeMount{
							kubev1.VolumeMount{
								Name:      "default-token-ztq95",
								ReadOnly:  true,
								MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								SubPath:   "",
							},
						},
						VolumeDevices:            []kubev1.VolumeDevice{},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						ImagePullPolicy:          "IfNotPresent",
						Stdin:                    false,
						StdinOnce:                false,
						TTY:                      false,
					},
				},
				RestartPolicy: "Always",
				DNSPolicy:     "ClusterFirst",
				NodeSelector:  map[string]string{},
			},
		},
	}
	var customPods kubernetes.PodList
	for _, p := range pods {
		// set service account
		p.Spec.ServiceAccountName = p.Name
		p.Spec.Containers = append(p.Spec.Containers, kubev1.Container{
			Name:       "linkerd-proxy",
			Image:      "gcr.io/linkerd-io/proxy:stable-2.3.2",
			Command:    []string{},
			Args:       []string{},
			WorkingDir: "",
			Ports: []kubev1.ContainerPort{
				kubev1.ContainerPort{
					Name:          "linkerd-proxy",
					HostPort:      0,
					ContainerPort: 4143,
					Protocol:      "TCP",
					HostIP:        "",
				},
				kubev1.ContainerPort{
					Name:          "linkerd-admin",
					HostPort:      0,
					ContainerPort: 4191,
					Protocol:      "TCP",
					HostIP:        "",
				},
			},
			EnvFrom: []kubev1.EnvFromSource{},
			Env: []kubev1.EnvVar{
				kubev1.EnvVar{
					Name:      "LINKERD2_PROXY_LOG",
					Value:     "warn,linkerd2_proxy=info",
					ValueFrom: (*kubev1.EnvVarSource)(nil),
				},
				kubev1.EnvVar{
					Name:      "LINKERD2_PROXY_DESTINATION_SVC_ADDR",
					Value:     "linkerd-destination.linkerd.svc.cluster.local:8086",
					ValueFrom: (*kubev1.EnvVarSource)(nil),
				},
				kubev1.EnvVar{
					Name:      "LINKERD2_PROXY_CONTROL_LISTEN_ADDR",
					Value:     "0.0.0.0:4190",
					ValueFrom: (*kubev1.EnvVarSource)(nil),
				},
				kubev1.EnvVar{
					Name:      "LINKERD2_PROXY_ADMIN_LISTEN_ADDR",
					Value:     "0.0.0.0:4191",
					ValueFrom: (*kubev1.EnvVarSource)(nil),
				},
				kubev1.EnvVar{
					Name:      "LINKERD2_PROXY_OUTBOUND_LISTEN_ADDR",
					Value:     "127.0.0.1:4140",
					ValueFrom: (*kubev1.EnvVarSource)(nil),
				},
				kubev1.EnvVar{
					Name:      "LINKERD2_PROXY_INBOUND_LISTEN_ADDR",
					Value:     "0.0.0.0:4143",
					ValueFrom: (*kubev1.EnvVarSource)(nil),
				},
				kubev1.EnvVar{
					Name:      "LINKERD2_PROXY_DESTINATION_PROFILE_SUFFIXES",
					Value:     "svc.cluster.local.",
					ValueFrom: (*kubev1.EnvVarSource)(nil),
				},
				kubev1.EnvVar{
					Name:      "LINKERD2_PROXY_INBOUND_ACCEPT_KEEPALIVE",
					Value:     "10000ms",
					ValueFrom: (*kubev1.EnvVarSource)(nil),
				},
				kubev1.EnvVar{
					Name:      "LINKERD2_PROXY_OUTBOUND_CONNECT_KEEPALIVE",
					Value:     "10000ms",
					ValueFrom: (*kubev1.EnvVarSource)(nil),
				},
				kubev1.EnvVar{
					Name:  "_pod_ns",
					Value: "",
					ValueFrom: &kubev1.EnvVarSource{
						FieldRef: &kubev1.ObjectFieldSelector{
							APIVersion: "v1",
							FieldPath:  "metadata.namespace",
						},
						ResourceFieldRef: (*kubev1.ResourceFieldSelector)(nil),
						ConfigMapKeyRef:  (*kubev1.ConfigMapKeySelector)(nil),
						SecretKeyRef:     (*kubev1.SecretKeySelector)(nil),
					},
				},
				kubev1.EnvVar{
					Name:      "LINKERD2_PROXY_DESTINATION_CONTEXT",
					Value:     "ns:$(_pod_ns)",
					ValueFrom: (*kubev1.EnvVarSource)(nil),
				},
				kubev1.EnvVar{
					Name:      "LINKERD2_PROXY_IDENTITY_DIR",
					Value:     "/var/run/linkerd/identity/end-entity",
					ValueFrom: (*kubev1.EnvVarSource)(nil),
				},
				kubev1.EnvVar{
					Name:      "LINKERD2_PROXY_IDENTITY_TRUST_ANCHORS",
					Value:     "-----BEGIN CERTIFICATE-----\nMIIBgzCCASmgAwIBAgIBATAKBggqhkjOPQQDAjApMScwJQYDVQQDEx5pZGVudGl0\neS5saW5rZXJkLmNsdXN0ZXIubG9jYWwwHhcNMTkwNjExMTc1NTIzWhcNMjAwNjEw\nMTc1NTQzWjApMScwJQYDVQQDEx5pZGVudGl0eS5saW5rZXJkLmNsdXN0ZXIubG9j\nYWwwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAASU/yvY+kB/kv4a1muTygXyoKG6\nAhv7+WR3NEJIg3kiNX/6KAPIe6BpD96VmZC9Gh10k20NvIn01ky++YykvwiEo0Iw\nQDAOBgNVHQ8BAf8EBAMCAQYwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMC\nMA8GA1UdEwEB/wQFMAMBAf8wCgYIKoZIzj0EAwIDSAAwRQIhAJLEnIPJpYoHFrUI\nES74RWiheFPHhHnISB2i4jTLVZ4SAiAxBdcSTBORcsiaaANbqqS9SSuTG3fXkmn4\nCI+aXWHK1Q==\n-----END CERTIFICATE-----\n",
					ValueFrom: (*kubev1.EnvVarSource)(nil),
				},
				kubev1.EnvVar{
					Name:      "LINKERD2_PROXY_IDENTITY_TOKEN_FILE",
					Value:     "/var/run/secrets/kubernetes.io/serviceaccount/token",
					ValueFrom: (*kubev1.EnvVarSource)(nil),
				},
				kubev1.EnvVar{
					Name:      "LINKERD2_PROXY_IDENTITY_SVC_ADDR",
					Value:     "linkerd-identity.linkerd.svc.cluster.local:8080",
					ValueFrom: (*kubev1.EnvVarSource)(nil),
				},
				kubev1.EnvVar{
					Name:  "_pod_sa",
					Value: "",
					ValueFrom: &kubev1.EnvVarSource{
						FieldRef: &kubev1.ObjectFieldSelector{
							APIVersion: "v1",
							FieldPath:  "spec.serviceAccountName",
						},
						ResourceFieldRef: (*kubev1.ResourceFieldSelector)(nil),
						ConfigMapKeyRef:  (*kubev1.ConfigMapKeySelector)(nil),
						SecretKeyRef:     (*kubev1.SecretKeySelector)(nil),
					},
				},
				kubev1.EnvVar{
					Name:      "_l5d_ns",
					Value:     "linkerd",
					ValueFrom: (*kubev1.EnvVarSource)(nil),
				},
				kubev1.EnvVar{
					Name:      "_l5d_trustdomain",
					Value:     "cluster.local",
					ValueFrom: (*kubev1.EnvVarSource)(nil),
				},
				kubev1.EnvVar{
					Name:      "LINKERD2_PROXY_IDENTITY_LOCAL_NAME",
					Value:     "$(_pod_sa).$(_pod_ns).serviceaccount.identity.$(_l5d_ns).$(_l5d_trustdomain)",
					ValueFrom: (*kubev1.EnvVarSource)(nil),
				},
				kubev1.EnvVar{
					Name:      "LINKERD2_PROXY_IDENTITY_SVC_NAME",
					Value:     "linkerd-identity.$(_l5d_ns).serviceaccount.identity.$(_l5d_ns).$(_l5d_trustdomain)",
					ValueFrom: (*kubev1.EnvVarSource)(nil),
				},
				kubev1.EnvVar{
					Name:      "LINKERD2_PROXY_DESTINATION_SVC_NAME",
					Value:     "linkerd-controller.$(_l5d_ns).serviceaccount.identity.$(_l5d_ns).$(_l5d_trustdomain)",
					ValueFrom: (*kubev1.EnvVarSource)(nil),
				},
			}})

		convertedPod := pod.FromKubePod(&p)

		customPods = append(customPods, convertedPod)
	}
	return customPods
}
