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
$ conftest gatekeeper -h
$ conftest gatekeeper --parameters example/parameters.yaml example/k8s_app.yaml -- --policy example/policy/ --all-namespaces


Testing CronJob - hello:
FAIL - - k8spspreadonlyrootfilesystem - only read-only root filesystem container is allowed: hello
FAIL - - k8spspallowprivilegeescalationcontainer - Privilege escalation container is not allowed: hello
FAIL - - capabilities - container <hello> is not dropping all required capabilities. Container must drop all of ["must_drop"]

7 tests, 4 passed, 0 warnings, 3 failures, 0 exceptions

Testing Job - hello:
FAIL - - k8spspreadonlyrootfilesystem - only read-only root filesystem container is allowed: hello
FAIL - - k8spspallowprivilegeescalationcontainer - Privilege escalation container is not allowed: hello
FAIL - - capabilities - container <hello> is not dropping all required capabilities. Container must drop all of ["must_drop"]

7 tests, 4 passed, 0 warnings, 3 failures, 0 exceptions

Testing Deployment - nginx-deployment:
FAIL - - k8spspreadonlyrootfilesystem - only read-only root filesystem container is allowed: nginx
FAIL - - k8spspallowprivilegeescalationcontainer - Privilege escalation container is not allowed: nginx
FAIL - - capabilities - container <nginx> is not dropping all required capabilities. Container must drop all of ["must_drop"]

7 tests, 4 passed, 0 warnings, 3 failures, 0 exceptions

Testing ReplicaSet - nginx-deployment:
FAIL - - k8spspreadonlyrootfilesystem - only read-only root filesystem container is allowed: nginx
FAIL - - k8spspallowprivilegeescalationcontainer - Privilege escalation container is not allowed: nginx
FAIL - - capabilities - container <nginx> is not dropping all required capabilities. Container must drop all of ["must_drop"]

7 tests, 4 passed, 0 warnings, 3 failures, 0 exceptions

Testing DaemonSet - fluentd-elasticsearch:
FAIL - - k8spspallowprivilegeescalationcontainer - Privilege escalation container is not allowed: fluentd-elasticsearch
FAIL - - capabilities - container <fluentd-elasticsearch> is not dropping all required capabilities. Container must drop all of ["must_drop"]
FAIL - - k8spspprivileged - Privileged container is not allowed: fluentd-elasticsearch, securityContext: {"allowPrivilegeEscalation": true, "privileged": true, "readOnlyRootFilesystem": false, "runAsNonRoot": false, "runAsUser": 0}
FAIL - - k8spspreadonlyrootfilesystem - only read-only root filesystem container is allowed: fluentd-elasticsearch

7 tests, 3 passed, 0 warnings, 4 failures, 0 exceptions

Testing Service - my-service:

7 tests, 7 passed, 0 warnings, 0 failures, 0 exceptions
Error: execute plugin: plugin exec: exit status 3
```
