# Gatekeeper Conftest plugin

A [Conftest](https://conftest.dev/) plugin that tranforms input objetcs to be compatible with [gatekeeper](https://github.com/open-policy-agent/gatekeeper) policies for pod admission.

The gatekeeper plugin tranforms the `input` object to the `input.review.object` object excepted by gatekeeper rego policies.\
The plugin also extracts the pod spec from a template where applicable (Deployment, DaemonSet, etc).

## Installation

The [releases](https://github.com/clover/gatekeeper-conftest/releases) page provides archive files containing the built binary and plugin.yaml file.\
Install the plugin using the built-in plugin manager by providing the URL to the release artifact.\
For example, installing on macOS:

```
conftest plugin install https://github.com/clover/gatekeeper-conftest/releases/download/v0.1.0/gatekeeper-conftest_0.1.0_Darwin_x86_64.tar.gz

```

Installing from source:

```
$ git clone https://github.com/clover/gatekeeper-conftest.git
$ cd gatekeeper-conftest
$ go build
$ conftest plugin install ./
```


## Usage

The gatekeeper plugin accepts a single kubernetes manifest and an optional file (--parameters) containing values for parameters used in constraint template rego policy.\
Any flags defined after a double dash `--` are passed to `conftest test` directly along with the transformed input.

```
$ conftest gatekeeper [flags] k8s_manifest_file.yaml -- [flags to pass to `conftest test`]

Flags:
  -h, --help                help for gatekeeper
  -p, --parameters string   path to file conatining paramater values used in constraint templates
```

### Exit codes
Exit code of 0: No policy failures or warnings.\
Exit code of 3: At least one policy failure.\
Exit code of 5: Plugin failure.

### Exit codes when using the `--fail-on-warn` conftest flag
Exit code of 0: No policy failures or warnings.\
Exit code of 3: No policy failures, but there exists at least one warning.\
Exit code of 4: At least one policy failure.\
Exit code of 5: Plugin failure.


## Example

```
$ conftest gatekeeper --parameters example/parameters.yaml example/k8s_app.yaml -- --policy example/policy/ --all-namespaces --no-fail
FAIL - .gatekeeper-conftest/CronJob_hello_pods.json - k8spspallowprivilegeescalationcontainer - Privilege escalation container is not allowed: hello
FAIL - .gatekeeper-conftest/DaemonSet_fluentd-elasticsearch_pods.json - k8spspallowprivilegeescalationcontainer - Privilege escalation container is not allowed: fluentd-elasticsearch
FAIL - .gatekeeper-conftest/Job_hello_pods.json - k8spspallowprivilegeescalationcontainer - Privilege escalation container is not allowed: hello
FAIL - .gatekeeper-conftest/ReplicaSet_nginx-deployment_pods.json - k8spspallowprivilegeescalationcontainer - Privilege escalation container is not allowed: nginx
FAIL - .gatekeeper-conftest/Deployment_nginx-deployment_pods.json - k8spspallowprivilegeescalationcontainer - Privilege escalation container is not allowed: nginx
FAIL - .gatekeeper-conftest/Deployment_nginx-deployment_pods.json - capabilities - container <nginx> is not dropping all required capabilities. Container must drop all of ["must_drop"]
FAIL - .gatekeeper-conftest/Job_hello_pods.json - capabilities - container <hello> is not dropping all required capabilities. Container must drop all of ["must_drop"]
FAIL - .gatekeeper-conftest/ReplicaSet_nginx-deployment_pods.json - capabilities - container <nginx> is not dropping all required capabilities. Container must drop all of ["must_drop"]
FAIL - .gatekeeper-conftest/CronJob_hello_pods.json - capabilities - container <hello> is not dropping all required capabilities. Container must drop all of ["must_drop"]
FAIL - .gatekeeper-conftest/DaemonSet_fluentd-elasticsearch_pods.json - capabilities - container <fluentd-elasticsearch> is not dropping all required capabilities. Container must drop all of ["must_drop"]
FAIL - .gatekeeper-conftest/DaemonSet_fluentd-elasticsearch_pods.json - k8spspprivileged - Privileged container is not allowed: fluentd-elasticsearch, securityContext: {"allowPrivilegeEscalation": true, "privileged": true, "readOnlyRootFilesystem": false, "runAsNonRoot": false, "runAsUser": 0}
FAIL - .gatekeeper-conftest/Deployment_nginx-deployment_pods.json - k8spspreadonlyrootfilesystem - only read-only root filesystem container is allowed: nginx
FAIL - .gatekeeper-conftest/Job_hello_pods.json - k8spspreadonlyrootfilesystem - only read-only root filesystem container is allowed: hello
FAIL - .gatekeeper-conftest/ReplicaSet_nginx-deployment_pods.json - k8spspreadonlyrootfilesystem - only read-only root filesystem container is allowed: nginx
FAIL - .gatekeeper-conftest/DaemonSet_fluentd-elasticsearch_pods.json - k8spspreadonlyrootfilesystem - only read-only root filesystem container is allowed: fluentd-elasticsearch
FAIL - .gatekeeper-conftest/CronJob_hello_pods.json - k8spspreadonlyrootfilesystem - only read-only root filesystem container is allowed: hello

77 tests, 61 passed, 0 warnings, 16 failures, 0 exceptions
```
