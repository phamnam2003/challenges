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

## Environment Variables

-

## Github Runner
