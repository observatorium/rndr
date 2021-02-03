package hellosvc

import golang "github.com/observatorium/rndr/examples/hellosvc/api/go"

// YOLO: Might be good direction, but guessing a lof for now. Need to go through references
// to ensure couple of things (how to create loops, merge defaults, use "self" data.
deployment: [_values=golang.#HelloService]: { // Kind of... need something like this??? (:
	apiVersion: "apps/v1"
	kind:       "Deployment"
	metadata: {
		labels: _values.CommonLabels
		name:      _values.Name
		namespace: _values.Namespace
	}
	spec: {
		replicas: _values.CommonLabels
		selector: matchLabels: _values.PodLabelSelector
		template: {
			metadata: labels: _values.CommonLabels
			spec: {
				affinity: podAntiAffinity: preferredDuringSchedulingIgnoredDuringExecution: [{
					podAffinityTerm: {
						labelSelector: matchExpressions: [{
							key:      "app.kubernetes.io/name"
							operator: "In"
							values: [
								"hellosvc", // Want [deployment.metadata.labels['app.kubernetes.io/name']],
							]
						}]
						namespaces: [
							_values.Namespace,
						]
						topologyKey: "kubernetes.io/hostname"
					}
					weight: 100
				}]
				containers: [{
					image: "paulbouwer/hello-kubernetes:" + _values.Version
					livenessProbe: {
						failureThreshold: 4
						httpGet: {
							path:   "/-/healthy"
							port:   _values.Ports[0].HTTP
							scheme: "HTTP"
						}
						periodSeconds: 30
					}
					name: _values.Name
					ports: [{
						containerPort: 80
						name:          "http"
					}] // How to loop over values.Ports?
					readinessProbe: {
						failureThreshold: 20
						httpGet: {
							path:   "/-/ready"
							port:   _values.Ports[0].HTTP
							scheme: "HTTP"
						}
						periodSeconds: 5
					}
					resources: {}
					terminationMessagePolicy: "FallbackToLogsOnError"
				}]
				terminationGracePeriodSeconds: 1
			}
		}
	}
}
