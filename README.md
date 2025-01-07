<p align="center">
<img src=".github/assets/logo.png" alt="BackHub Logo" width="250" height="250" /><br>
<h1 align="center">BackHub</h1><br>

<p align="center">
<a href="https://github.com/tanq16/backhub/actions/workflows/release.yml"><img src="https://github.com/tanq16/backhub/actions/workflows/release.yml/badge.svg" alt="Release Build"></a>&nbsp;<a href="https://goreportcard.com/report/github.com/tanq16/backhub"><img src="https://goreportcard.com/badge/github.com/tanq16/backhub" alt="Go Report Card"></a><br>
<a href="https://github.com/Tanq16/backhub/releases"><img alt="GitHub Release" src="https://img.shields.io/github/v/release/tanq16/backhub"></a>&nbsp;<a href="https://hub.docker.com/r/tanq16/backhub"><img alt="Docker Pulls" src="https://img.shields.io/docker/pulls/tanq16/backhub"></a>
</p>
</p>

`BackHub` is a simple yet powerful GitHub repository backup tool that creates and maintains complete local mirrors of your repositories. It supports concurrent backups, automated scheduling (every 3 days), and can be run either as a standalone binary or as a Docker container.

---

# Features

- Full repository mirroring including all branches, tags, and history
- Concurrent backup processing for multiple repositories defined in a YAML config file
- Automated backups with Docker deployment (every 3 days)
- GitHub token-based authentication (to be used in an environment variable)
- Easy restoration capability due to it being local mirror
- Multi-arch and multi-OS binary for simple one-time usage

# Installation

### Go Install

```bash
go install github.com/tanq16/backhub@latest
```

### Docker Installation

```bash
mkdir $HOME/backhub # this is where you put your .backhub.yaml file
```

```bash
docker run -d \
  --name backhub \
  -e GH_TOKEN=your_github_token \
  -v $HOME/backhub:/app \
  tanq16/backhub:latest
```

### Building from Source

```bash
git clone https://github.com/tanq16/backhub.git && cd backhub
go build
```

# Usage

### Binary Mode

Run directly with default config path:
```bash
backhub
```

Specify custom config path:
```bash
backhub -c /path/to/config.yaml
```

For inline environment variable, use as:

```bash
GH_TOKEN=pat_jsdhksjdskhjdhkajshkdjh backhub
```

### Docker Mode

The Docker container automatically runs backups every 3 days.

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

BackHub uses a simple YAML configuration file. Default path is `.backhub.yaml`:

```yaml
repos:
  - github.com/username/repo1
  - github.com/username/repo2
  - github.com/org/repo3
```

Lastly, use the `GH_TOKEN` environment variable as your GitHub personal access token to use perform the backup.

# Using the Local Mirrors

To use a local mirror as a Git repository source (like when you need to restore from the backup), the following can be done:

1. Directly pull or clone from the mirror:
    ```bash
    # Add the mirror as remote to an existing repo
    git remote add backup /path/to/your/mirror.git
    git pull backup main # or any other branch

    # clone from the mirror
    git clone /path/to/your/mirror.git new-repo
    ```

2. Serve the mirror as a local Git server:
    ```bash
    # In the mirror directory
    git daemon --base-path=/path/to/parent --export-all
    # Clone from the new git server
    git clone git://localhost/mirror.git
    ```

3. Using file protocol in the Git URL:
    ```bash
    git clone file:///path/to/mirror.git
    ```

Since it's a mirror, it contains all refs (branches, tags, etc.), so when you clone or pull from it, it can access everything like from the original. Use the following to see all branches and tags in the mirror:

```bash
git branch -a  # shows all branches
git tag -l     # shows all tags
```
