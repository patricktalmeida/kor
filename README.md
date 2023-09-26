![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/yonahd/kor)
![GitHub release (with filter)](https://img.shields.io/github/v/release/yonahd/kor?color=green&link=https://github.com/yonahd/kor/releases)
[![codecov](https://codecov.io/gh/yonahd/kor/branch/main/graph/badge.svg?token=tNKcOjlxLo)](https://codecov.io/gh/yonahd/kor)

# Kor - Kubernetes Orphaned Resources Finder

Kor is a tool to discover unused Kubernetes resources. Currently, Kor can identify and list unused:
- ConfigMaps  
- Secrets
- Services
- ServiceAccounts
- Deployments
- StatefulSets
- Roles
- HPAs
- PVCs
- Ingresses
- PDBs

![Kor Screenshot](/images/screenshot.png)

## Installation

Download the binary for your operating system from the [releases page](https://github.com/yonahd/kor/releases) and add it to your system's PATH.

For macOS users, you can install Kor using Homebrew:
```sh
brew install kor
```

## Usage

Kor provides various subcommands to identify and list unused resources. The available commands are:

- `all` - Gets all unused resources for the specified namespace or all namespaces.
- `configmap` - Gets unused ConfigMaps for the specified namespace or all namespaces.
- `secret` - Gets unused Secrets for the specified namespace or all namespaces.
- `services` - Gets unused Services for the specified namespace or all namespaces.
- `serviceaccount` - Gets unused ServiceAccounts for the specified namespace or all namespaces.
- `deployments` - Gets unused Deployments for the specified namespace or all namespaces.
- `statefulsets` - Gets unused StatefulSets for the specified namespace or all namespaces.
- `role` - Gets unused Roles for the specified namespace or all namespaces.
- `hpa` - Gets unused HPAs for the specified namespace or all namespaces.
- `pvc` - Gets unused PVCs for the specified namespace or all namespaces.
- `ingress` - Gets unused Ingresses for the specified namespace or all namespaces.
- `pdb` - Gets unused PDBs for the specified namespace or all namespaces.

### Supported Flags
```
-e, --exclude-namespaces string   Namespaces to be excluded, split by comma. Example: --exclude-namespace ns1,ns2,ns3. If --include-namespace is set, --exclude-namespaces will be ignored.
-h, --help                        help for kor
-n, --include-namespaces string   Namespaces to run on, split by comma. Example: --include-namespace ns1,ns2,ns3.
-k, --kubeconfig string           Path to kubeconfig file (optional)
    --output string               Output format (table or json) (default "table")
    --slack-auth-token string     Slack auth token to send notifications to. --slack-auth-token requires --slack-channel to be set.
    --slack-channel string        Slack channel to send notifications to. --slack-channel requires --slack-auth-token to be set.
    --slack-webhook-url string    Slack webhook URL to send notifications to
```

To use a specific subcommand, run `kor [subcommand] [flags]`.

```sh
kor all --namespace my-namespace
```

For more information about each subcommand and its available flags, you can use the `--help` flag.

```sh
kor [subcommand] --help
```

## Supported resources and limitations

| Resource        | What it looks for                                                                                                                                                                                                                  | Known False Positives  ⚠️                                                                                                     |
|-----------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------|
| ConfigMaps      | ConfigMaps not used in the following places:<br/>- Pods<br/>- Containers<br/>- ConfigMaps used through Volumes<br/>- ConfigMaps used through environment variables                                                                 | ConfigMaps used by resources which don't explicitly state them in the config.<br/> e.g Grafana dashboards loaded dynamically OPA policies fluentd configs |
| Secrets         | Secrets not used in the following places:<br/>- Pods<br/>- Containers<br/>- Secrets used through volumes<br/>- Secrets used through environment variables<br/>- Secrets used by Ingress TLS<br/>- Secrets used by ServiceAccounts |    Secrets used by resources which don't explicitly state them in the config                                                                                                                         |
| Services        | Services with no endpoints                                                                                                                                                                                                         |                                                                                                                              |
| Deployments     | Deployments with no Replicas                                                                                                                                                                                                       |                                                                                                                              |
| ServiceAccounts | ServiceAccounts unused by Pods<br/>ServiceAccounts unused by roleBinding or clusterRoleBinding                                                                                                                                     |                                                                                                                              |
| StatefulSets    | Statefulsets with no Replicas                                                                                                                                                                                                      |                                                                                                                              |
| Roles           | Roles not used in roleBinding                                                                                                                                                                                                      |                                                                                                                              |
| PVCs            | PVCs not used in Pods                                                                                                                                                                                                              |                                                                                                                              |
| Ingresses       | Ingresses not pointing at any Service                                                                                                                                                                                              |                                                                                                                              |
| Hpas            | HPAs not used in Deployments<br/> HPAs not used in StatefulSets                                                                                                                                                                    |                                                                                                                              |
| Pdbs            | PDBs not used in Deployments<br/> PDBs not used in StatefulSets                                                                                                                                                                    |                                                                                                                              |


## Ignore Resources
The resources labeled with "kor/used = true" will be ignored by kor even if they are unused. You can add this label to resources you want to ignore.

## Import Option
You can also use kor as a Go library to programmatically discover unused resources. By importing the github.com/yonahd/kor/pkg/kor package, you can call the relevant functions to retrieve unused resources. The library provides the option to get the results in JSON format by specifying the outputFormat parameter.

```go
import (
    "github.com/yonahd/kor/pkg/kor"
)

func main() {
    myNamespaces := kor.IncludeExcludeLists{
        IncludeListStr: "my-namespace1, my-namespace2",
    }
    outputFormat := "json" // Set to "json" for JSON output

    if outputFormat == "json" {
        jsonResponse, err := kor.GetUnusedDeploymentsStructured(myNamespaces, kubeconfig, "json")
        if err != nil {
            // Handle error
        }
        // Process the JSON response
        // ...
    } else {
        kor.GetUnusedDeployments(namespace)
    }
}
```

## In Cluster Usage

To use this tool inside the cluster running as a CronJob and sending the results to a Slack Webhook as raw text(has characters limits of 4000) or to a Slack channel by uploading a file(recommended), you can use the following commands:

```sh
# Send to a Slack webhook as raw text
helm upgrade -i kor \
    --namespace kor \
    --create-namespace \
    --set cronJob.slackWebhookUrl=<slack-webhook-url> \
    ./charts/kor
```

```sh
# Send to a Slack channel by uploading a file
helm upgrade -i kor \
    --namespace kor \
    --create-namespace \
    --set cronJob.slackChannel=<slack-channel> \
    --set cronJob.slackToken=<slack-token> \
    ./charts/kor
```
> Note: To send it to Slack as a file it's required to set the `slackToken` and `slackChannel` values.

It's set to run every Monday at 1 a.m. by default. You can change the schedule by setting the `cronJob.schedule` value.

```sh
helm upgrade -i kor \
    --namespace kor \
    --create-namespace \
    --set cronJob.slackChannel=<slack-channel> \
    --set cronJob.slackToken=<slack-token> \
    --set cronJob.schedule="0 1 * * 1" \
    ./charts/kor
```

## Contributing

Contributions are welcome! If you encounter any bugs or have suggestions for improvements, please open an issue in the [issue tracker](https://github.com/yonahd/kor/issues).

## License

This open-source project is available under the [MIT License](LICENSE). Feel free to use, modify, and distribute it as per the terms of the license.

