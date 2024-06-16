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
