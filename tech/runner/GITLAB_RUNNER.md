# RUNNER

- `GitLab Runner` is an application that works with `GitLab CI/CD` to run jobs in a pipeline.
- When developers push code to `GitLab`, they can define automated tasks in a `.gitlab-ci.yml` file. These tasks might include *running tests*, *building applications*, or *deploying code*. `GitLab Runner` is the application that executes these tasks on computing infrastructure.
- As an administrator, you are responsible for providing and managing the infrastructure where these CI/CD jobs run. This involves installing `GitLab Runner` applications, configuring them, and ensuring they have adequate capacity to handle your organization’s CI/CD workload.
- Runner registration is the process that links the runner with one or more GitLab instances. You must register the runner so that it can pick up jobs from the GitLab instance. [Register](https://docs.gitlab.com/runner/register/)

## Installation

- [Installation and Configuration Guide](https://docs.gitlab.com/runner/install/)
- You can install `GitLab Runner` on various operating systems, including Linux, Windows, and macOS. You can also run it in a Docker container or use it as a Kubernetes pod via package managers, binaries, or Docker images.

```bash
# Ubuntu/Debian/Mint
curl -L "https://packages.gitlab.com/install/repositories/runner/gitlab-runner/script.deb.sh" | sudo bash
sudo apt install gitlab-runner

# RHEL/CentOS/Fedora/Amazon Linux
sudo yum install gitlab-runner

or

sudo dnf install gitlab-runnerurl -L "https://packages.gitlab.com/install/repositories/runner/gitlab-runner/script.rpm.sh" | sudo bash
```

- You can install `Gitlab Runner` using Docker with [image](https://hub.docker.com/r/gitlab/gitlab-runner)

## Registration

- `Gitlab Runner` must be registered with a `GitLab instance` to pick up and *execute jobs*. During registration, you provide details such as the `GitLab` instance URL, a registration token, and the executor type (e.g., shell, Docker, Kubernetes).

```bash
sudo gitlab-runner register
```

- Enter the `Gitlab` instance URL (ex: `https://gitlab.com`, `http://192.168.32.21`, etc.). After that, enter the registration token you obtained from your `GitLab` instance. To get the token, navigate to your project or group in `GitLab`, go to `Settings` > `CI/CD` > `Runners`, and get command for `gitlab runner`.

```bash
gitlab-runner register  --url https://gitlab.com  --token <TOKEN_GITLAB_RUNNER>
```

- Start the runner:

```bash
sudo gitlab-runner run
```

> [!Note]
> You can change the executor type during registration. Common options include `shell`, `docker`, and `kubernetes`. Choose the one that best fits your infrastructure and job requirements.
> You can create gitlab-runner for self-hosted gitlab instance or gitlab.com. Often in companies, self-hosted gitlab instance is used
