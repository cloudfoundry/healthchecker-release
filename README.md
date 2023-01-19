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

1. Vendor the healthchecker BOSH package in your boshrelease.
1. Add `<job>-healthchecker` process to your job's monit file, based on our [example monit file](jobs/example/monit).
1. Add BPM config for the healthchecker, based on our [example bpm.yml](jobs/example/templates/bpm.yml.erb).
1. Add a healthchecker config template to your job based on our [example helathchecker config template](jobs/example/templates/healthchecker.yml.erb).
1. Update the job's spec to add the healthchecker config template, healthchecker package,
   and any new properties.
