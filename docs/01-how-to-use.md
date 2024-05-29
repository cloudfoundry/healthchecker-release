---
title: How to use
expires_at: never
tags: [healthchecker-release]
---

<!-- vim-markdown-toc GFM -->

  * [How to use](#how-to-use)
* [Implementing in Your BOSH Release](#implementing-in-your-bosh-release)
    * [1. Vendor the `healthchecker` package](#1-vendor-the-healthchecker-package)
    * [2. Edit Monit File](#2-edit-monit-file)
    * [3. Add BPM Config](#3-add-bpm-config)
    * [4. Add Healthchecker config file](#4-add-healthchecker-config-file)
    * [5. Update job spec](#5-update-job-spec)

<!-- vim-markdown-toc -->

### How to use

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
Add `{JOB}-healthchecker` process to the desired job's monit file, based on the `example-healthchecker` process in our [example monit file](../jobs/example/monit#L7-L13).

1. Copy the `example-healthchecker` process to your monit file.
1. Change all occurances of `example` to `{JOB}`, e.g. `example-healthchecker` to `{JOB}-healthchecker` and `example-healthchecker.pid` to `{JOB}-healthchecker.pid`
1. The path `/var/vcap/packages/healthchecker/bin/restart-monit-job` *MUST* match the healthchecker PREFIX. Update the path appropriately with the PREFIX, e.g. `/var/vcap/packages/{PREFIX}-healthchecker/bin/restart-monit-job`.

#### 3. Add BPM Config
Add BPM config for the new healthchecker process, based on our [example bpm.yml](../jobs/example/templates/bpm.yml.erb#L7-L14).

1. Copy the `example-healthchecker` process to your BPM file.
1. Change all occurances of `example` to `{JOB}`, e.g. `example-healthchecker` to `{JOB}-healthchecker` and `/var/vcap/jobs/example/config/healthchecker.yml` to `/var/vcap/jobs/{JOB}/config/healthchecker.yml`
1. The path `/var/vcap/packages/healthchecker/bin/healthchecker` *MUST* match the healthchecker PREFIX. Update the path appropriately with the PREFIX, e.g. `/var/vcap/packages/{PREFIX}-healthchecker/bin/healthchecker`.

#### 4. Add Healthchecker config file
Add a healthchecker config template to your job based on our [example healthchecker config template](../jobs/example/templates/healthchecker.yml.erb).

1. Copy the config file to `{YOUR_RELEASE}/jobs/{JOB}/templates/healthchecker.yml.erb`.
1. Update the `component_name` to match your job, e.g. `{JOB}-healthchecker`.

#### 5. Update job spec
Update the job's spec to add the healthchecker config template, healthchecker package, and any new properties.

Based on the example, you need to add:
- [healthchecker config template](../jobs/example/spec#L6)
- [package itself](../jobs/example/spec#L10)
  - *Important*: this *MUST* match the PREFIX used when vendoring, i.e. `{PREFIX}-healthchecker`
- [`failure_counter_file` property](../jobs/example/spec#L25-L27)
