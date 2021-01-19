// No need to define defaults here - rndr will allow to have consistent defaults defined in proto or go.
local defaults = {
  local defaults = self,
  name: 'example',
  namespace: 'default',
  version: '1.8',
  replicas: 1,
  resources: {},
  ports: {
    http: 80,
  },
  message: 'hello',
  tracing: {},

  commonLabels:: {
    'app.kubernetes.io/name': 'hellosvc',
    'app.kubernetes.io/instance': defaults.name,
    'app.kubernetes.io/version': defaults.version,
    'app.kubernetes.io/component': 'demo',
  },

  podLabelSelector:: {
    [labelName]: defaults.commonLabels[labelName]
    for labelName in std.objectFields(defaults.commonLabels)
    if labelName != 'app.kubernetes.io/version'
  },
};

// values definition is availabile in ../api/
// TODO(bwplotka): Generate validation and safety check if this file is using correct field names etc from go/protobuf definition.
function(values) {
  local hs = self,

  config:: defaults + values,

  // Safety checks for combined config of defaults and params.
  assert std.isNumber(hs.config.replicas) && hs.config.replicas >= 0 : 'hello pod replicas has to be number >= 0',
  assert std.isObject(hs.config.resources),

  service: {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      name: hs.config.name,
      namespace: hs.config.namespace,
      labels: hs.config.commonLabels,
    },
    spec: {
      ports: [
        {
          assert std.isString(name),
          assert std.isNumber(hs.config.ports[name]),

          name: name,
          port: hs.config.ports[name],
          targetPort: hs.config.ports[name],
        }
        for name in std.objectFields(hs.config.ports)
      ],
      selector: hs.config.podLabelSelector,
    },
  },

  deployment:
    local c = {
      name: hs.config.name,
      image: 'paulbouwer/hello-kubernetes:%s' % hs.config.version,
      ports: [
        { name: port.name, containerPort: port.port }
        for port in hs.service.spec.ports
      ],
      livenessProbe: { failureThreshold: 4, periodSeconds: 30, httpGet: {
        scheme: 'HTTP',
        port: hs.service.spec.ports[0].port,
        path: '/-/healthy',
      } },
      readinessProbe: { failureThreshold: 20, periodSeconds: 5, httpGet: {
        scheme: 'HTTP',
        port: hs.service.spec.ports[0].port,
        path: '/-/ready',
      } },
      resources: if hs.config.resources != {} then hs.config.resources else {},
      terminationMessagePolicy: 'FallbackToLogsOnError',
    };

    {
      apiVersion: 'apps/v1',
      kind: 'Deployment',
      metadata: {
        name: hs.config.name,
        namespace: hs.config.namespace,
        labels: hs.config.commonLabels,
      },
      spec: {
        replicas: hs.config.replicas,
        selector: { matchLabels: hs.config.podLabelSelector },
        template: {
          metadata: {
            labels: hs.config.commonLabels,
          },
          spec: {
            containers: [c],
            terminationGracePeriodSeconds: 1,
            affinity: { podAntiAffinity: {
              preferredDuringSchedulingIgnoredDuringExecution: [{
                podAffinityTerm: {
                  namespaces: [hs.config.namespace],
                  topologyKey: 'kubernetes.io/hostname',
                  labelSelector: { matchExpressions: [{
                    key: 'app.kubernetes.io/name',
                    operator: 'In',
                    values: [hs.deployment.metadata.labels['app.kubernetes.io/name']],
                  }] },
                },
                weight: 100,
              }],
            } },
          },
        },
      },
    },
}
