package hellosvc

deployment: example: {
	apiVersion: "apps/v1"
	kind:       "Deployment"
	metadata: {
		labels: {
			"app.kubernetes.io/component": "demo"
			"app.kubernetes.io/instance":  "example"
			"app.kubernetes.io/name":      "hellosvc"
			"app.kubernetes.io/version":   "1.8"
		}
		name:      "example"
		namespace: "default"
	}
	spec: {
		replicas: 1
		selector: matchLabels: {
			"app.kubernetes.io/component": "demo"
			"app.kubernetes.io/instance":  "example"
			"app.kubernetes.io/name":      "hellosvc"
		}
		template: {
			metadata: labels: {
				"app.kubernetes.io/component": "demo"
				"app.kubernetes.io/instance":  "example"
				"app.kubernetes.io/name":      "hellosvc"
				"app.kubernetes.io/version":   "1.8"
			}
			spec: {
				affinity: podAntiAffinity: preferredDuringSchedulingIgnoredDuringExecution: [{
					podAffinityTerm: {
						labelSelector: matchExpressions: [{
							key:      "app.kubernetes.io/name"
							operator: "In"
							values: [
								"hellosvc",
							]
						}]
						namespaces: [
							"default",
						]
						topologyKey: "kubernetes.io/hostname"
					}
					weight: 100
				}]
				containers: [{
					image: "paulbouwer/hello-kubernetes:1.8"
					livenessProbe: {
						failureThreshold: 4
						httpGet: {
							path:   "/-/healthy"
							port:   80
							scheme: "HTTP"
						}
						periodSeconds: 30
					}
					name: "example"
					ports: [{
						containerPort: 80
						name:          "http"
					}]
					readinessProbe: {
						failureThreshold: 20
						httpGet: {
							path:   "/-/ready"
							port:   80
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
