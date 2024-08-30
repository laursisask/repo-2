# konnectivity-agent helm chart

This Helm chart deploys the Konnectivity Agent, a component of the Kubernetes Konnectivity service.

## prerequisites

- Kubernetes 1.18+
- Helm 3.0+

## installing the chart

To install the chart with the release name `my-release`:

```bash
helm install my-release ./konnectivity-agent
```

This command deploys the Konnectivity Agent on the Kubernetes cluster with the default configuration.

## uninstalling the chart

To uninstall/delete the `my-release` deployment:

```bash
helm delete my-release
```

This command removes all the Kubernetes components associated with the chart and deletes the release.

## configuration

The following table lists the configurable parameters of the Konnectivity Agent chart and their default values.

| Parameter | Description | Default |
| --------- | ----------- | ------- |
| `image.repository` | Konnectivity Agent image repository | `registry.k8s.io/kas-network-proxy/proxy-agent` |
| `image.tag` | Konnectivity Agent image tag | `v0.0.37` |
| `proxyServer.host` | The host of the proxy server | `""` |
| `proxyServer.port` | The port of the proxy server | `8132` |
| `adminServer.port` | The port of the admin server | `8133` |
| `healthServer.port` | The port of the health server | `8134` |

You can specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.
