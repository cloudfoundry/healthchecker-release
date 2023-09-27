## healthchecker

`healthchecker` is a go executable designed to perform TCP/HTTP based health checks of
processes managed by `monit` in BOSH releases. Since the version of `monit` included in
BOSH does not support specific tcp/http health checks, we designed this utility to perform
health checking and restart processes if they become unreachable.

### How it Works

`healthchecker` is added to a boshrelease as a `monit` process under the Job that is to be monitored.
It is configured to perform a healthcheck against the main process in the Job. If `healthchecker`
detects a failure, it will panic and exit. `monit` should be configured to run the `restart-monit-job`
script on the failure of the `healthchecker` process. This script restarts the main monit process,
up to ten failures in a row. After 10 consecutive failures, it gives up, since restarting the process
is either in a horrible state, or the `healthchecker` is misconfigured and should not be causing
process downtime.

## Implementing in Your BOSH Release

#### 1. Vendor the `healthchecker` package
Vendor the `healthchecker` BOSH package in your bosh release using a PREFIX. This will name the `healthchecker` package `{PREFIX}-healthchecker` in your release to prevent collisions.
```
git clone https://github.com/cloudfoundry/healthchecker-release.git
cd ~/PATH/YOUR_RELEASE
bosh vendor-package healthchecker ~/workspace/healthchecker-release --prefix ${PREFIX}
```

The PREFIX can be anything but should probably be the name of your release or a specific job in the release that will use healthchecker, e.g. `cf-networking` to make `cf-networking-healthchecker`.
Once the `healthchecker` package has been vendored into your release, any job in your release can make use of it by following the rest of the instructions.

#### 2. Edit Monit File
Add `{JOB}-healthchecker` process to the desired job's monit file, based on the `example-healthchecker` process in our [example monit file](jobs/example/monit#L7-L13).

1. Copy the `example-healthchecker` process to your monit file.
1. Change all occurances of `example` to `{JOB}`, e.g. `example-healthchecker` to `{JOB}-healthchecker` and `example-healthchecker.pid` to `{JOB}-healthchecker.pid`
1. The path `/var/vcap/packages/healthchecker/bin/restart-monit-job` *MUST* match the healthchecker PREFIX. Update the path appropriately with the PREFIX, e.g. `/var/vcap/packages/{PREFIX}-healthchecker/bin/restart-monit-job`.

#### 3. Add BPM Config
Add BPM config for the new healthchecker process, based on our [example bpm.yml](jobs/example/templates/bpm.yml.erb#L7-L14).

1. Copy the `example-healthchecker` process to your BPM file.
1. Change all occurances of `example` to `{JOB}`, e.g. `example-healthchecker` to `{JOB}-healthchecker` and `/var/vcap/jobs/example/config/healthchecker.yml` to `/var/vcap/jobs/{JOB}/config/healthchecker.yml`
1. The path `/var/vcap/packages/healthchecker/bin/healthchecker` *MUST* match the healthchecker PREFIX. Update the path appropriately with the PREFIX, e.g. `/var/vcap/packages/{PREFIX}-healthchecker/bin/healthchecker`.

#### 4. Add Healthchecker config file
Add a healthchecker config template to your job based on our [example healthchecker config template](jobs/example/templates/healthchecker.yml.erb).

1. Copy the config file to `{YOUR_RELEASE}/jobs/{JOB}/templates/healthchecker.yml.erb`.
1. Update the `component_name` to match your job, e.g. `{JOB}-healthchecker`.

#### 5. Update job spec
Update the job's spec to add the healthchecker config template, healthchecker package, and any new properties.

Based on the example, you need to add:
- [healthchecker config template](jobs/example/spec#L6)
- [package itself](jobs/example/spec#L10)
  - *Important*: this *MUST* match the PREFIX used when vendoring, i.e. `{PREFIX}-healthchecker`
- [`failure_counter_file` property](jobs/example/spec#L25-L27)

#### <a name="running-unit-and-integration-tests"></a> Running Unit and Integration Tests

##### With Docker

Running tests for this release

- `./scripts/create-docker-container.bash`: This will create a docker container with appropriate mounts.
- `./scripts/test-in-docker-locally.bash`: Create docker container and run all tests and setup in a single script.
  - `./scripts/test-in-docker-locally.bash <package> <sub-package>`: For running tests under a specific package and/or sub-package: e.g. `./scripts/test-in-docker-locally.bash healthchcker watchdog`

When inside docker container: 
- `/repo/scripts/docker/test.bash`: This will run all tests in this release
- `/repo/scripts/docker/test.bash healthchcker`: This will only run `healthchckers` tests
- `/repo/scripts/docker/test.bash healthchcker integration`: This will only run `watchdog` sub-package tests for `healthchcker` package
- `/repo/scripts/docker/tests-templates.bash`: This will run all of tests for bosh tempalates
- `/repo/scripts/docker/lint.bash`: This will run all of linting defined for this repo.
