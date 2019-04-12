package inputs

import (
	"github.com/gogo/protobuf/proto"
	kubernetes2 "github.com/solo-io/supergloo/pkg/api/custom/clients/kubernetes"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BookInfoPods(svcNamespace string) v1.PodList {
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
	var customPods v1.PodList
	for _, p := range pods {
		// set service account
		p.Spec.ServiceAccountName = p.Name

		convertedPod, err := kubernetes2.FromKube(&p)
		if err != nil {
			panic(err)
		}
		customPods = append(customPods, convertedPod)
	}
	return customPods
}
