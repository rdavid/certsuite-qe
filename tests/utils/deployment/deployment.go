package deployment

import (
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

type MultusAnnotation struct {
	Name string `json:"name"`
}

// DefineDeployment returns deployment struct.
func DefineDeployment(deploymentName string, namespace string, image string, label map[string]string) *v1.Deployment {
	return &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace},
		Spec: v1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: label,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "testpod-",
					Labels: label,
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: pointer.Int64Ptr(0),
					Containers: []corev1.Container{
						{
							Name:    "test",
							Image:   image,
							Command: []string{"/bin/bash", "-c", "sleep INF"},
						},
					},
				},
			},
		},
	}
}

// RedefineAllContainersWithPreStopSpec redefines deployment with requested lifecycle/preStop spec.
func RedefineAllContainersWithPreStopSpec(deployment *v1.Deployment, command []string) {
	for index := range deployment.Spec.Template.Spec.Containers {
		deployment.Spec.Template.Spec.Containers[index].Lifecycle = &corev1.Lifecycle{
			PreStop: &corev1.LifecycleHandler{
				Exec: &corev1.ExecAction{
					Command: command,
				},
			},
		}
	}
}

// RedefineWithContainersSecurityContextAll redefines deployment with extended permissions.
func RedefineWithContainersSecurityContextAll(deployment *v1.Deployment) {
	for index := range deployment.Spec.Template.Spec.Containers {
		deployment.Spec.Template.Spec.Containers[index].SecurityContext = &corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{"ALL"}},
		}
	}
}

// RedefineWithLabels redefines deployment with additional label.
func RedefineWithLabels(deployment *v1.Deployment, label map[string]string) {
	newMap := make(map[string]string)
	for k, v := range deployment.Spec.Template.Labels {
		newMap[k] = v
	}

	for k, v := range label {
		newMap[k] = v
	}

	deployment.Spec.Template.Labels = newMap
}

// RedefineWithMultus redefines deployment with additional labels.
func RedefineWithMultus(deployment *v1.Deployment, nadNames []string) *v1.Deployment {
	if len(nadNames) < 1 {
		return deployment
	}

	var nadAnnotations []MultusAnnotation

	for _, nadName := range nadNames {
		nadAnnotations = append(nadAnnotations, MultusAnnotation{Name: nadName})
	}

	bString, _ := json.Marshal(nadAnnotations)

	deployment.Spec.Template.Annotations = map[string]string{
		"k8s.v1.cni.cncf.io/networks": string(bString),
	}

	return deployment
}

// RedefineWithReplicaNumber redefines deployment with requested replica number.
func RedefineWithReplicaNumber(deployment *v1.Deployment, replicasNumber int32) {
	deployment.Spec.Replicas = pointer.Int32Ptr(replicasNumber)
}

// RedefineFirstContainerWithPreStopSpec redefines deployment first container with lifecycle/preStop spec.
func RedefineFirstContainerWithPreStopSpec(deployment *v1.Deployment, command []string) error {
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		deployment.Spec.Template.Spec.Containers[0].Lifecycle = &corev1.Lifecycle{
			PreStop: &corev1.LifecycleHandler{
				Exec: &corev1.ExecAction{
					Command: command}}}

		return nil
	}

	return fmt.Errorf("deployment %s does not have any containers", deployment.Name)
}

// RedefineWithTerminationGracePeriod redefines deployment with terminationGracePeriod spec.
func RedefineWithTerminationGracePeriod(deployment *v1.Deployment, terminationGracePeriod *int64) {
	deployment.Spec.Template.Spec.TerminationGracePeriodSeconds = terminationGracePeriod
}

// RedefineWithPodAntiAffinity redefines deployment with podAntiAffinity spec.
func RedefineWithPodAntiAffinity(deployment *v1.Deployment, label map[string]string) {
	deployment.Spec.Template.Spec.Affinity = &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: label,
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			},
		}}
}

func RedefineWithImagePullPolicy(deployment *v1.Deployment, pullPolicy corev1.PullPolicy) {
	for index := range deployment.Spec.Template.Spec.Containers {
		deployment.Spec.Template.Spec.Containers[index].ImagePullPolicy = pullPolicy
	}
}

func RedefineWithNodeSelector(deployment *v1.Deployment, nodeSelector map[string]string) {
	deployment.Spec.Template.Spec.NodeSelector = nodeSelector
}

func RedefineWithNodeAffinity(deployment *v1.Deployment, key string) {
	deployment.Spec.Template.Spec.Affinity = &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      key,
								Operator: corev1.NodeSelectorOpExists,
							},
						},
					},
				},
			},
		}}
}

func RedefineWithReadinessProbe(deployment *v1.Deployment) {
	for index := range deployment.Spec.Template.Spec.Containers {
		deployment.Spec.Template.Spec.Containers[index].ReadinessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"ls"},
				},
			},
		}
	}
}

func RedefineWithLivenessProbe(deployment *v1.Deployment) {
	for index := range deployment.Spec.Template.Spec.Containers {
		deployment.Spec.Template.Spec.Containers[index].LivenessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"ls"},
				},
			},
		}
	}
}

// RedefineWithStartUpProbe adds startup probe to deployment manifest.
func RedefineWithStartUpProbe(deployment *v1.Deployment) {
	for index := range deployment.Spec.Template.Spec.Containers {
		deployment.Spec.Template.Spec.Containers[index].StartupProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"ls"},
				},
			},
		}
	}
}

func RedefineWithContainerSpecs(deployment *v1.Deployment, containerSpecs []corev1.Container) {
	deployment.Spec.Template.Spec.Containers = containerSpecs
}

func RedefineWithPriviledgedContainer(deployment *v1.Deployment) {
	for index := range deployment.Spec.Template.Spec.Containers {
		deployment.Spec.Template.Spec.Containers[index].SecurityContext = &corev1.SecurityContext{
			Privileged: pointer.Bool(true),
			RunAsUser:  pointer.Int64(0),
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{"ALL"}},
		}
	}
}

func RedefineWithHostPid(deployment *v1.Deployment, hostPid bool) {
	deployment.Spec.Template.Spec.HostPID = hostPid
}

func RedefineWithHostIpc(deployment *v1.Deployment, hostIpc bool) {
	deployment.Spec.Template.Spec.HostIPC = hostIpc
}

func RedefineWithAutomountServiceAccountToken(deployment *v1.Deployment, token bool) {
	deployment.Spec.Template.Spec.AutomountServiceAccountToken = &token
}

func RedefineWithHostNetwork(deployment *v1.Deployment, hostNetwork bool) {
	deployment.Spec.Template.Spec.HostNetwork = hostNetwork
}

func RedefineWithPVC(deployment *v1.Deployment, name string, claim string) {
	deployment.Spec.Template.Spec.Volumes = []corev1.Volume{
		{
			Name: name,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: claim,
				},
			},
		},
	}
}

func RedefineWithCPUResources(deployment *v1.Deployment, limit string, req string) {
	for i := range deployment.Spec.Template.Spec.Containers {
		deployment.Spec.Template.Spec.Containers[i].Resources = corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse(limit),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse(req),
			},
		}
	}
}

func RedefineWithRunTimeClass(deployment *v1.Deployment, rtcName string) {
	deployment.Spec.Template.Spec.RuntimeClassName = pointer.String(rtcName)
}

func RedefineWithShareProcessNamespace(deployment *v1.Deployment, shareProcessNamespace bool) {
	deployment.Spec.Template.Spec.ShareProcessNamespace = &shareProcessNamespace
}

func RedefineWithSysPtrace(deployment *v1.Deployment) {
	for index := range deployment.Spec.Template.Spec.Containers {
		deployment.Spec.Template.Spec.Containers[index].SecurityContext = &corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{"SYS_PTRACE"}},
		}
	}
}
