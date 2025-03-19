<p align="center">
<img src=".github/assets/logo.png" alt="BackHub Logo" width="275" /><br>
<h1 align="center">BackHub</h1><br>

<p align="center">
<a href="https://github.com/tanq16/backhub/actions/workflows/release.yml"><img src="https://github.com/tanq16/backhub/actions/workflows/release.yml/badge.svg" alt="Release Build"></a><br>
<a href="https://github.com/Tanq16/backhub/releases"><img alt="GitHub Release" src="https://img.shields.io/github/v/release/tanq16/backhub"></a>&nbsp;<a href="https://hub.docker.com/r/tanq16/backhub"><img alt="Docker Pulls" src="https://img.shields.io/docker/pulls/tanq16/backhub"></a>
</p>
</p>

<p align="center">
<b>BackHub</b> is a simple GitHub repository backup tool that creates complete local mirrors of your repositories. It supports concurrent backup operations, automated scheduling (every 3 days when self-hosted with Docker), and offers a standalone binary as well.
</p>

---

# Features

- Full repository mirroring including all branches, tags, and history
- Concurrent backup processing for multiple repositories defined in a YAML config file
- Automated backups and homelab-deployable with Docker deployment (every 3 days)
- GitHub token-based authentication (to be used in an environment variable)
- Easy restoration capability due to it being local mirror
- Multi-arch and multi-OS binary for simple one-time usage

# Installation

The easiest way to download it is from the [project releases](https://github.com/Tanq16/backhub/releases).

### Go Install

```bash
go install github.com/tanq16/backhub@latest
```

### Building from Source

```bash
git clone https://github.com/tanq16/backhub.git && \
cd backhub && \
go build
```

# Usage

### Binary Mode

Backhub uses an environment variable to authenticate to GitHub. To do this, set your `GH_TOKEN` variable. This can be done inline with:

```bash
GH_TOKEN=pat_jsdhksjdskhjdhkajshkdjh backhub
```

With the environment variable in the picture, run `backhub` by directly specifying a repository or specify a custom config like so:

```bash
# config file
backhub /path/to/config.yaml

# direct repo
backhub github.com/tanq16/backhub
```

### Docker Mode

The Docker container uses a script that automatically runs the tool every 3 days to provide scheduled backups. First, set up a persistence repository:

```bash
mkdir $HOME/backhub # this is where the mirrors will be stored
```

Then run the container like so:

```bash
docker run -d \
  --name backhub \
  -e GH_TOKEN=your_github_token \
  -v $HOME/backhub:/app \
  tanq16/backhub:latest
```

Conversely, you can use the following compose file for Docker compose or a stack in Portainer, Dockge, etc.

```yaml
version: "3.8"
services:
  backhub:
    image: tanq16/backhub:latest
    restart: unless-stopped
    environment:
      - GH_TOKEN=your_github_token
    volumes:
      - /home/tanq/backhub:/app
```

# YAML Config File

BackHub uses a simple YAML configuration file:

```yaml
repos:
  - github.com/username/repo1
  - github.com/username/repo2
  - github.com/org/repo3
```

For Docker, put the config file in the mounted directory and name it `config.yaml`.

# Using the Local Mirrors

To use a local mirror as a Git repository source (like when you need to restore from the backup), the following can be done:

1. Directly pull or clone from the mirror treating it as a `backup` remote in an existing repository:
    ```bash
    git remote add backup /path/to/your/mirror.git
    git pull backup main # or any other branch
    git clone /path/to/your/mirror.git new-repo
    ```
2. Serve the mirror via a local git server and use it :
    ```bash
    git daemon --base-path=/path/to/mirror --export-all
    git clone git://localhost/mirror.git # in a different path
    ```
3. Use the mirror as a git server by refering to it through the file protocol:
    ```bash
    git clone file:///path/to/mirror.git
    ```

Being a mirror, it contains all references (branches, tags, etc.), so cloning or pulling from it allows accessing everything as if it's the original. Use `git branch -a` to see all branches and `git tag -l` to see all tags in the mirror.
