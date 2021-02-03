package golang

import (
	corev1 "k8s.io/api/core/v1"
)

// Default returns default values for HelloService config.
func Default() HelloService {
	h := HelloService{
		Name:      "example",
		Version:   "1.8",
		Namespace: "default",
		Replicas:  1,
		Ports:     Ports{HTTP: 80},
		Message:   "hello",
	}

	h.PodLabelSelector = map[string]string{
		"app.kubernetes.io/name":      "hellosvc",
		"app.kubernetes.io/instance":  h.Name,
		"app.kubernetes.io/component": "demo",
	}
	h.CommonLabels = map[string]string{
		"app.kubernetes.io/version": h.Version,
	}

	for k, v := range h.PodLabelSelector {
		h.CommonLabels[k] = v
	}
	return h
}

type HelloService struct {
	Name      string
	Namespace string
	Version   string
	Replicas  int
	Resources corev1.ResourceRequirements
	Ports     Ports
	// TODO(bwplotka): With Go we could play in having '+' prefix telling to add values only. Experiment with this.
	CommonLabels     map[string]string
	PodLabelSelector map[string]string

	Message string

	// Extra allows to provide raw bytes in renderer specific language allowing adhoc
	// adjustments right before resources generation allowing quick adjustments.
	// Use on your own responsibility.
	Extra []byte
}

type Ports struct {
	HTTP int
}
