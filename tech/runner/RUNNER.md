# Github action

- `GitHub Actions` is a *continuous integration* and *continuous delivery* (CI/CD) platform that allows you to automate your *build*, *test*, and *deployment pipeline*. You can create *workflows* that run tests whenever you push a change to your repository, or that deploy *merged pull requests* to production.

## Components

- You can configure a GitHub Actions workflow to be triggered when an event occurs in your repository, such as a *pull request* being opened or an issue being created. Your workflow contains one or more jobs which can run in sequential order or in parallel. Each job will run inside its own *virtual machine runner*, or *inside a container*, and has one or more steps that either run a script that you define or run an action, which is a reusable extension that can simplify your workflow.

### Workflows

- A `workflow` is a configurable *automated process* that will run *one or more jobs*. Workflows are defined by a `YAML` file checked in to your repository and will run when triggered by an event in your repository, or they can be triggered manually, or at a defined schedule.
- Workflows are defined in the `.github/workflows` directory in a repository. A repository can have multiple workflows, each of which can perform a different set of tasks such as:
  - Building and testing pull requests
  - Deploying your application every time a release is created
  - Adding a label whenever a new issue is opened

### Events

- An `event` is a *specific activity* in a repository that triggers a **workflow** run. For example, an activity can originate from GitHub when someone creates a pull request, opens an issue, or pushes a commit to a repository. You can also trigger a workflow to run on a [schedule](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#schedule), by [posting to a REST API](https://docs.github.com/en/rest/repos/repos#create-a-repository-dispatch-event), or manually.

### Jobs

- A `job` is *a set of steps* in a `workflow` that is executed on the same `runner`. Each step is either a shell script that will be executed, or an `action` that will be run. Steps are executed in order and are dependent on each other. Since each step is executed on the same `runner`, you can share data from one step to another. For example, you can have a step that builds your application followed by a step that tests the application that was built.
- You can configure a job's dependencies with other jobs; by default, jobs have no dependencies and run in parallel. When a job takes a dependency on another job, it waits for the dependent job to complete before running.

### Actions

- An `action` is a *pre-defined*, *reusable* set of jobs or code that *performs specific tasks* within a `workflow`, reducing the amount of repetitive code you write in your workflow files. Actions can perform tasks such as:
  - Pulling your Git repository from GitHub
  - Setting up the correct `toolchain` for your build environment
  - Setting up authentication to your `cloud provider`
- You can write your own actions, or you can find actions to use in your workflows in the GitHub Marketplace.

### Runners

- A `runner` is a **server** that runs your `workflows` when they're triggered. Each `runner` can run a single job at a time. GitHub provides *Ubuntu Linux*, *Microsoft Windows*, and *macOS* runners to run your workflows. Each workflow run executes in a fresh, newly-provisioned virtual machine.

## Concepts

### Variables

- `Variables` provide a way to store and reuse non-sensitive configuration information. You can store any configuration data such as `compiler flags`, `usernames`, or `server names` as variables. Variables are `interpolated` on the runner machine that runs your `workflow`. Commands that run in actions or workflow steps can *create*, *read*, and *modify variables*.

```yaml
name: Custom Variables
on: push

env:
  GLOBAL_VAR: "I am a global variable"

jobs:
  set_env:
    runs-on: ubuntu-latest
    steps:
      - name: Set a job-level variable
        run: echo "JOB_VAR=I am a job variable" >> $GITHUB_ENV

      - name: Use the variables
        run: |
          echo $GLOBAL_VAR
          echo $JOB_VAR
      - name: "Assigning output to a variable"
        run: echo "$CHANNEL_NAME"
        env:
          CHNANNEL_NAME: "stable"
```

- [Default Variables](https://docs.github.com/en/actions/reference/workflows-and-actions/variables)
- Secret Variables:
  - `Secret variables` are encrypted environment variables created in a repository and can only be used by GitHub Actions. You can create secrets in a repository, organization, or environment. Secrets are not passed to workflows that are triggered by a pull request from a forked repository.
  - You can create secrets in the `Settings` tab of a repository, organization, or environment. Secrets created in a repository are only available to that repository. Secrets created in an organization are available to all repositories in the organization. Secrets created in an environment are only available to workflows that reference the environment.

```yaml
steps:
  - name: Use secret variable
    run: echo ${{ secrets.MY_SECRET }}
```

### Decrypt and Encrypt files

- To use secrets that are larger than `64KB` in your workflow, you can use the `actions/checkout` action to check out your repository, and then use the `gpg` command to decrypt the file. You can then use the file in your workflow, and when you're done, you can use the `gpg` command to encrypt the file again.

### Contexts

- `Contexts` are a way to access information about workflow runs, variables, runner environments, jobs, and steps. Each context is an object that contains properties, which can be strings or other objects.

```yaml
# ${{ <context> }}

name: CI
on: push
jobs:
  prod-check:
    if: ${{ github.ref == 'refs/heads/main' }}
    runs-on: ubuntu-latest
    steps:
      - run: echo "Deploying to production server on branch $GITHUB_REF"
```
