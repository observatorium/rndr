package hellosvc

service: example: {
	apiVersion: "v1"
	kind:       "Service"
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
		ports: [{
			name:       "http"
			port:       80
			targetPort: 80
		}]
		selector: {
			"app.kubernetes.io/component": "demo"
			"app.kubernetes.io/instance":  "example"
			"app.kubernetes.io/name":      "hellosvc"
		}
		type: "LoadBalancer"
	}
}
