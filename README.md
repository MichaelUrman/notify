# notify/teams
Post GitHub workflow events to a Microsoft Teams webhook

Configuration:
- Create your incoming webhook in Microsoft Teams.
- Optionally store the webhook URL in a secret (e.g. MSTEAMS_NOTIFY_HOOK_URL)
- Run only after ensuring Go is on the Runner (uses: actions/setup-go)
- Create your workflow yaml, running on any scenarios you want to notify about (See [Workflow syntax](https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#on) and [Events that trigger workflows](https://docs.github.com/en/actions/reference/events-that-trigger-workflows))
- Specify the secret or webhook URL in your workflow configuration:

```yaml
on:
  push:
  pull_request:
  release:
  check_run:
  create:
  delete:
  public:
  pull_request:
  pull_request_review:
  pull_request_review_comment:

jobs:
  notify:
    name: "Notify Microsoft Teams"
    steps:
    - uses: actions/setup-go@v2
      with:
        go-version: '~1.14'
    # ...
    - uses: MichaelUrman/notify/teams@tip # or use a sha
      with:
        hookurl: ${{ secrets.MSTEAMS_NOTIFY_HOOK_URL }}
```

For reporting CI results, add a step after your CI step with `if: always()` and a `job-status` like this:
```
jobs:
  ci_test:
    name "Integration Tests"
    steps:
    - uses: actions/setup-go@v2
      with:
        go-version: '~1.14'

    # step here that runs integration tests

    - uses: MichaelUrman/notify/teams@tip
      if: always()
      with:
        hookurl: ${{ secrets.MSTEAMS_NOTIFY_HOOK_URL }}
        job-status: ${{ job.status }}
```
Alternately, you can report a specific step's outcome as the job-status:
```
      with:
        hookurl: ${{ secrets.MSTEAMS_NOTIFY_HOOK_URL }}
        job-status: ${{ steps.stepname.outcome }}
```
