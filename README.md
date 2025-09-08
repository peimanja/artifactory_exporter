# JFrog Artifactory Exporter

[![Go Build](https://github.com/peimanja/artifactory_exporter/actions/workflows/release.yml/badge.svg)](https://github.com/peimanja/artifactory_exporter/actions/workflows/release.yml) [![Publish Image](https://github.com/peimanja/artifactory_exporter/actions/workflows/publish-image.yml/badge.svg)](https://github.com/peimanja/artifactory_exporter/actions/workflows/publish-image.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/peimanja/artifactory_exporter)](https://goreportcard.com/report/github.com/peimanja/artifactory_exporter)

A [Prometheus](https://prometheus.io) exporter for [JFrog Artifactory](https://jfrog.com/artifactory) stats.


## Note

This exporter is under development and more metrics will be added later. Tested on Artifactory Commercial, Enterprise and OSS version `6.x` and `7.x`.

## Authentication

The Artifactory exporter requires **admin** user and it supports multiple means of authentication. The following methods are supported:
  * Basic Auth
  * Bearer Token

### Basic Auth

Basic auth may be used by setting `ARTI_USERNAME` and `ARTI_PASSWORD` environment variables.

### Bearer Token

Artifactory access tokens may be used via the Authorization header by setting `ARTI_ACCESS_TOKEN` environment variable.

## Usage

### Binary

Download the binary for your operating system from [release](https://github.com/peimanja/artifactory_exporter/releases) page and run it:
```bash
$ ./artifactory_exporter <flags>
```

### Docker

Set the credentials in `env_file_name` and you can deploy this exporter using the [peimanja/artifactory_exporter](https://registry.hub.docker.com/r/peimanja/artifactory_exporter/) Docker image:

```bash
$ docker run --env-file=env_file_name -p 9531:9531 peimanja/artifactory_exporter:latest <flags>
```

### Caching

#### Docker Compose

Running the exporter against an Artifactory instance with millions of artifacts will cause performance issues in case Prometheus will scrape too often.

To avoid such situations you can run nginx in front of the exporter and use it as a cache. The Artifactory responses will be cached in nginx and kept valid for `PROXY_CACHE_VALID` seconds. After that time any new request from Prometheus will re-request metrics from Artifactory and store again in the nginx cache.

Set the credentials in an environment file as described in the *Docker* section and store the file as `.env` next to `docker-compose.yml` and run the following command:

```bash
docker-compose up -d
```

#### Artifactory API response cache

On large artifactory clusters, the response times for certain API calls can be very long, which can lead to timeouts when scraping metrics.
To avoid this, you can enable caching of API responses by setting the `--use-cache` flag. This will cache successful API responses for a specified time (`--cache-ttl`) and use them for subsequent requests that exceed the specified timeout (`--cache-timeout`).


## Install with Helm

[Helm](https://helm.sh) must be installed to use the charts.
Please refer to Helm's [documentation](https://helm.sh/docs/) to get started.

Source code for the exporter helm chart can be found [here peimanja/helm-charts](https://github.com/peimanja/helm-charts/tree/main/charts/prometheus-artifactory-exporter)

Once Helm is set up properly, add the repo as follows:

### Prerequisites

- Kubernetes 1.8+ with Beta APIs enabled

### Add Repo

```console
helm repo add peimanja https://peimanja.github.io/helm-charts
helm repo update
```

_See [helm repo](https://helm.sh/docs/helm/helm_repo/) for command documentation._

### Configuration

See [Customizing the Chart Before Installing](https://helm.sh/docs/intro/using_helm/#customizing-the-chart-before-installing). To see all configurable options with detailed comments, visit the chart's [values.yaml](https://github.com/peimanja/helm-charts/blob/main/charts/prometheus-artifactory-exporter/values.yaml), or run these configuration commands:


```console
# Helm 3
helm show values peimanja/prometheus-artifactory-exporter
```

Set your values in `myvals.yaml`:
```yaml
artifactory:
  url: http://artifactory:8081/artifactory
  accessToken: "xxxxxxxxxxxxxxxxxxxx"
  existingSecret: false

options:
  logLevel: info
  logFormat: logfmt
  telemetryPath: /metrics
  verifySSL: false
  timeout: 5s
  optionalMetrics:
    - replication_status
    - federation_status
```

### Install Chart

```console
# Helm 3
helm install -f myvals.yaml [RELEASE_NAME] peimanja/prometheus-artifactory-exporter
```

_See [configuration](#configuration) below._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

### Flags

```bash
$  docker run peimanja/artifactory_exporter:latest -h
usage: artifactory_exporter [<flags>]

Flags:
  -h, --help                    Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=":9531"
                                Address to listen on for web interface and telemetry.
      --web.telemetry-path="/metrics"
                                Path under which to expose metrics.
      --artifactory.scrape-uri="http://localhost:8081/artifactory"
                                URI on which to scrape JFrog Artifactory.
      --artifactory.ssl-verify  Flag that enables SSL certificate verification for the scrape URI
      --artifactory.timeout=5s  Timeout for trying to get stats from JFrog Artifactory.
      --access-federation-target=ACCESS-FEDERATION-TARGET
                                URL of JFrog Access Federation Target server. Only required if optional metric AccessFederationValidate is enabled
      --use-cache               Use cache for API responses to circumvent timeouts
      --cache-timeout=30s       Timeout for API responses to fallback to cache
      --cache-ttl=5m            Time to live for cached API responses
      --artifacts-time-interval=1m... ...
                                Time interval for created and downloaded stats
      --optional-metric=metric-name ...
                                optional metric to be enabled. Valid metrics are: [artifacts replication_status federation_status open_metrics access_federation_validate background_tasks]. Pass multiple times to enable multiple optional metrics.
      --log.level=info          Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt       Output format of log messages. One of: [logfmt, json]
      --version                 Show application version.
```

| Flag / Environment Variable                    | Required | Default                             | Description                                                                                                                                                                              |
|------------------------------------------------|----------|-------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `web.listen-address`<br/>`WEB_LISTEN_ADDR`     | No       | `:9531`                             | Address to listen on for web interface and telemetry.                                                                                                                                    |
| `web.telemetry-path`<br/>`WEB_TELEMETRY_PATH`  | No       | `/metrics`                          | Path under which to expose metrics.                                                                                                                                                      |
| `artifactory.scrape-uri`<br/>`ARTI_SCRAPE_URI` | No       | `http://localhost:8081/artifactory` | URI on which to scrape JFrog Artifactory.                                                                                                                                                |
| `artifactory.ssl-verify`<br/>`ARTI_SSL_VERIFY` | No       | `true`                              | Flag that enables SSL certificate verification for the scrape URI.                                                                                                                       |
| `artifactory.timeout`<br/>`ARTI_TIMEOUT`       | No       | `5s`                                | Timeout for trying to get stats from JFrog Artifactory.                                                                                                                                  |
| `use-cache`<br/>`USE_CACHE`                    | No       | `false`                             | Use caching for API responses to circumvent timeouts.                                                                                                                                    |
| `cache-timeout`<br/>`CACHE_TIMEOUT`            | No       | `30s`                               | Timeout for API responses before falling back to cache. Requires enabling `use-cache` to apply this. Should be set to a lower value than `artifactory.timeout` to reap caching benefits. |
| `cache-ttl`<br/>`CACHE_TTL`                    | No       | `5m`                                | Time to live for cached API responses. Requires enabling `use-cache` to apply this.                                                                                                      |
| `artifacts-time-interval`                      | No       | `1m`,`5m`, `15m`                    | Time interval for created and downloaded stats. Requires enabling `--optional-metric metrics` to apply this.                                                                             |
| `optional-metric`                              | No       |                                     | optional metric to be enabled. Pass multiple times to enable multiple optional metrics.                                                                                                  |
| `log.level`                                    | No       | `info`                              | Only log messages with the given severity or above. One of: [debug, info, warn, error].                                                                                                  |
| `log.format`                                   | No       | `logfmt`                            | Output format of log messages. One of: [logfmt, json].                                                                                                                                   |
| `ARTI_USERNAME`                                | *No      |                                     | User to access Artifactory                                                                                                                                                               |
| `ARTI_PASSWORD`                                | *No      |                                     | Password of the user accessing the Artifactory                                                                                                                                           |
| `ARTI_ACCESS_TOKEN`                            | *No      |                                     | Access token for accessing the Artifactory                                                                                                                                               |

* Either `ARTI_USERNAME` and `ARTI_PASSWORD` or `ARTI_ACCESS_TOKEN` environment variables has to be set.

### Metrics

Some metrics are not available with Artifactory OSS license. The exporter returns the following metrics:

| Metric                                    | Description                                                               | Labels                                        | OSS support |
|-------------------------------------------|---------------------------------------------------------------------------|-----------------------------------------------|-------------|
| artifactory_up                            | Was the last scrape of Artifactory successful.                            |                                               | &#9989;     |
| artifactory_exporter_build_info           | Exporter build information.                                               | `version`, `revision`, `branch`, `goversion`  | &#9989;     |
| artifactory_exporter_total_scrapes        | Current total artifactory scrapes.                                        |                                               | &#9989;     |
| artifactory_exporter_total_api_errors     | Current total Artifactory API errors when scraping for stats.             |                                               | &#9989;     |
| artifactory_exporter_json_parse_failures  | Number of errors while parsing Json.                                      |                                               | &#9989;     |
| artifactory_replication_enabled           | Replication status for an Artifactory repository (1 = enabled).           | `name`, `type`, `cron_exp`, `status`          |             |
| artifactory_security_certificates         | SSL certificate name and expiry as labels, seconds to expiration as value | `alias`, `expires`, `issued_by`               |             |
| artifactory_security_groups               | Number of Artifactory groups.                                             |                                               |             |
| artifactory_security_users                | Number of Artifactory users for each realm.                               | `realm`                                       |             |
| artifactory_storage_artifacts             | Total artifacts count stored in Artifactory.                              |                                               | &#9989;     |
| artifactory_storage_artifacts_size_bytes  | Total artifacts Size stored in Artifactory in bytes.                      |                                               | &#9989;     |
| artifactory_storage_binaries              | Total binaries count stored in Artifactory.                               |                                               | &#9989;     |
| artifactory_storage_binaries_size_bytes   | Total binaries Size stored in Artifactory in bytes.                       |                                               | &#9989;     |
| artifactory_storage_filestore_bytes       | Total space in the file store in bytes.                                   | `storage_dir`, `storage_type`                 | &#9989;     |
| artifactory_storage_filestore_used_bytes  | Space used in the file store in bytes.                                    | `storage_dir`, `storage_type`                 | &#9989;     |
| artifactory_storage_filestore_free_bytes  | Space free in the file store in bytes.                                    | `storage_dir`, `storage_type`                 | &#9989;     |
| artifactory_storage_repo_used_bytes       | Space used by an Artifactory repository in bytes.                         | `name`, `package_type`, `type`                | &#9989;     |
| artifactory_storage_repo_folders          | Number of folders in an Artifactory repository.                           | `name`, `package_type`, `type`                | &#9989;     |
| artifactory_storage_repo_files            | Number of files in an Artifactory repository.                             | `name`, `package_type`, `type`                | &#9989;     |
| artifactory_storage_repo_items            | Number of items in an Artifactory repository.                             | `name`, `package_type`, `type`                | &#9989;     |
| artifactory_artifacts_created_1m          | Number of artifacts created in the repo (last 1 minute).                  | `name`, `package_type`, `type`                | &#9989;     |
| artifactory_artifacts_created_5m          | Number of artifacts created in the repo (last 5 minutes).                 | `name`, `package_type`, `type`                | &#9989;     |
| artifactory_artifacts_created_15m         | Number of artifacts created in the repo (last 15 minutes).                | `name`, `package_type`, `type`                | &#9989;     |
| artifactory_artifacts_downloaded_1m       | Number of artifacts downloaded from the repository (last 1 minute).       | `name`, `package_type`, `type`                | &#9989;     |
| artifactory_artifacts_downloaded_5m       | Number of artifacts downloaded from the repository (last 5 minutes).      | `name`, `package_type`, `type`                | &#9989;     |
| artifactory_artifacts_downloaded_15m      | Number of artifacts downloaded from the repository (last 15 minutes).     | `name`, `package_type`, `type`                | &#9989;     |
| artifactory_system_healthy                | Is Artifactory working properly (1 = healthy).                            |                                               | &#9989;     |
| artifactory_system_license                | License type and expiry as labels, seconds to expiration as value         | `type`, `licensed_to`, `expires`              | &#9989;     |
| artifactory_system_version                | Version and revision of Artifactory as labels.                            | `version`, `revision`                         | &#9989;     |
| artifactory_federation_mirror_lag         | Federation mirror lag in milliseconds.                                    | `name`, `remote_url`, `remote_name`           |             |
| artifactory_federation_unavailable_mirror | Unsynchronized federated mirror status.                                   | `status`, `name`, `remote_url`, `remote_name` |             |

* Common labels:
  * `node_id`: Artifactory node ID that the metric is scraped from.

#### Optional metrics

Some metrics are expensive to compute and are disabled by default. To enable them, use `--optional-metric=metric_name` flag. Use this with caution as it may impact the performance in Artifactory instances with many repositories.

Supported optional metrics:

* `artifacts` - Extracts number of artifacts created/downloaded for each repository. Enabling this will add `artifactory_artifacts_*` metrics. Please note that on large Artifactory instances, this may impact the performance.
* `replication_status` - Extracts status of replication for each repository which has replication enabled. Enabling this will add the `status` label to `artifactory_replication_enabled` metric.
* `federation_status` - Extracts federation metrics. Enabling this will add two new metrics: `artifactory_federation_mirror_lag`, and `artifactory_federation_unavailable_mirror`. Please note that these metrics are only available in Artifactory Enterprise Plus and version 7.18.3 and above.
* `open_metrics` - Exposes Open Metrics from the JFrog Platform. For more information about Open Metrics, please refer to [JFrog Platform Open Metrics](https://jfrog.com/help/r/jfrog-platform-administration-documentation/open-metrics).
* `access_federation_validate` - Validates whether trust is established towards a given JFrog Access Federation target server. Requires optional parameter `access-federation-target` to be set to the URL of the target server as well as token-based authentication. For more information, please refer to [JFrog Access Federation Circle of Trust validation](https://jfrog.com/help/r/jfrog-rest-apis/validate-target-for-circle-of-trust).
* `background_tasks` - Tracks the number of Artifactory background tasks by type and state. Enabling this will add the `artifactory_background_tasks` metric. Use this to monitor scheduled, running, stopped, or canceled tasks.

### Grafana Dashboard

Dashboard can be found [here](https://grafana.com/grafana/dashboards/12113).

![Grafana Dashboard](/grafana/dashboard-screenshot-1.png)
![Grafana Dashboard](/grafana/dashboard-screenshot-2.png)

### Common Issues

In most cases enabling debug logs will help to identify the issue. To enable debug logs, use `--log.level=debug` flag.

#### No metrics are being scraped

* Check if the exporter is running and listening on the port specified by `--web.listen-address` flag.
* Check if `artifactory_up` metric is `1` or `0`. If it is `0`, check the logs for the error message.
* Check if `artifactory_exporter_total_api_errors` metric is `0`. If it is not `0` and it is increasing, check the logs for the error message.

#### Some metrics or labels are missing

* Check the logs to see if there are any timeouts or errors while scraping for metrics. In a large Artifactory instance, it may take a long time to scrape for all metrics especially `artifactory_artifacts_*` metrics. If there are any errors, try increasing the default timeout(5s) using `--artifactory.timeout` flag.
* Some metrics are not available based on your version or license type. Check the [metrics](#metrics) section to see if the metric is available for your license type.
* Some metrics are optional and are disabled by default. Check the [optional metrics](#optional-metrics) section to see available optional metrics. You can enable them using `--optional-metric=metric_name` flag. You can pass this flag multiple times to enable multiple optional metrics.

#### There was an error when trying to unmarshal the API Error

This error usually indicates that the exporter is not able to properly reach the Artifactory endpoint. One of the common reasons for this is that the Artifactory URL is not set properly. Make sure you are not missing `/artifactory` at the end of the URL which is how most implementations of Artifactory are configured. (e.g. `http://artifactory.yourdomain.com/artifactory`)
