<p align="center">
<img src=".github/assets/logo.png" alt="BackHub Logo" width="250" />
</p>
<h1 align="center">BackHub</h1>
<p align="center">
<a href="https://github.com/tanq16/backhub/actions/workflows/release.yml"><img src="https://github.com/tanq16/backhub/actions/workflows/release.yml/badge.svg" alt="Release"></a>&nbsp;<a href="https://github.com/Tanq16/backhub/releases"><img alt="GitHub Release" src="https://img.shields.io/github/v/release/tanq16/backhub"></a>
</p>

<p align="center">
<b>BackHub</b> is a simple GitHub repository backup tool that creates complete local mirrors of your repositories.
</p>

---

# Features

- Full repository mirroring including all branches, tags, and history
- Concurrent backup processing for multiple repositories defined in a YAML config file
- GitHub token-based authentication (to be used in an environment variable)
- Easy restoration capability due to it being local mirror
- Multi-arch and multi-OS binary for simple one-time usage

# Installation

### Binary

The easiest way to use Backhub is to download it from the [project releases](https://github.com/Tanq16/backhub/releases) for your OS and architecture.

### Development Build

```bash
go install github.com/tanq16/backhub@latest
```

### Build from Source

```bash
git clone https://github.com/tanq16/backhub.git && \
cd backhub && \
go build
```

# Usage

Backhub uses an environment variable to authenticate to GitHub. To do this, set your `GH_TOKEN` variable. This can be done inline with:

```bash
GH_TOKEN=pat_jsdhksjdskhjdhkajshkdjh backhub
```

Alternatively, export it to your shell session with:

```bash
export GH_TOKEN=pat_jsdhksjdskhjdhkajshkdjh
```

With the environment variable exported, `backhub` can be directly executed multiple times from the command line like so:

```bash
# config file
backhub /path/to/config.yaml

# direct repo
backhub github.com/tanq16/backhub
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
