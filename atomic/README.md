# Fedora Atomic

A custom Fedora Atomic container image dagger module heavily customized for
my personal use.

> [!NOTE]
> Interested in making your own? Create an issue, we can work
> together to build a proper template.

Dagger is used to build and publish the image. The heavy lifting behind
Fedora Atomic images is done by the Fedora and Universal Blue communities.

- [Fedora Atomic](https://fedoraproject.org/atomic-desktops/)
- [Universal Blue](https://universal-blue.org)

## Usage

See the [GitHub Actions workflow](../.github/workflows/atomic.yaml) or
[`justfile`](justfile) for examples.

Useful commands:

```bash
just                         # print just recipes
dagger call -m atomic --help # print help for atomic Dagger module
```

## Install and Rebase

1. [Install Fedora Silverblue](https://docs.fedoraproject.org/en-US/fedora-silverblue/installation/)
2. Rebase to an image from this project

    ```bash
    # change IMAGE and TAG as desired
    # https://github.com/scottames?tab=packages&repo_name=containers
    IMAGE=atomic-silverblue-main \
    TAG=42 \
      rpm-ostree rebase \
        ostree-unverified-registry:ghcr.io/scottames/$IMAGE:$TAG
    ```

3. Rebase to the same image signed

    ```bash
    # change IMAGE and TAG as desired
    # https://github.com/scottames?tab=packages&repo_name=containers
    IMAGE=atomic-silverblue-main \
    TAG=42 \
    rpm-ostree rebase \
        ostree-image-signed:docker://ghcr.io/scottames/$IMAGE:$TAG
    ```
