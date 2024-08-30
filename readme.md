# charts
> Helm charts for UpCloud-specific components.

## usage

To use these charts, add the repository to Helm:

```bash
helm repo add upcloud https://upcloud.github.io/charts
```

Then, you can install the charts using the repository name:

```bash
helm install my-release upcloud/<chart-name>
```

## charts

- [konnectivity-agent](./konnectivity-agent) - Helm chart for the Konnectivity Agent, a component of the Kubernetes Konnectivity service.
