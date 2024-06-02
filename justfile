_default:
    @just --list --list-heading $'' --list-prefix $''

alias df := dagger-fedora

atomic-ublue-silverblue *FLAGS='container':
  dagger call -m atomic \
    --registry ghcr.io \
    --org      ublue-os \
    --tag      40 \
    --variant  silverblue \
    --suffix   main \
    {{FLAGS}}

atomic-ublue-silverblue-publish:
  dagger call -m atomic \
    --registry ghcr.io \
    --org      ublue-os \
    --tag      40 \
    --variant  silverblue \
    --suffix   main \
    --source . \
  publish \
		--registry=ghcr.io \
		--image-name=atomic-silverblue \
		--username=$GITHUB_USERNAME \
		--secret=$GITHUB_TOKEN \
		--tag=test


dagger-fedora *FLAGS='container':
  dagger call -m fedora \
    --registry ghcr.io \
    --org      ublue-os \
    --tag      40 \
    --variant  silverblue \
    --suffix   main \
    {{FLAGS}}

dagger-develop:
  cd fedora
  dagger develop --sdk=go
  cd custom
  dagger develop --sdk=go
