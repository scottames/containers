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
    --source   . \
    {{FLAGS}}

atomic-ublue-silverblue-publish-and-sign:
  dagger call -m atomic \
    --registry ghcr.io \
    --org      ublue-os \
    --tag      40 \
    --variant  silverblue \
    --suffix   main \
    --source . \
    publish \
  		--registry="ghcr.io/scottames" \
  		--image-name=atomic-silverblue \
  		--username=$GITHUB_USERNAME \
  		--secret=env:GITHUB_TOKEN \
  		--tags=test1,test2 \
    sign \
      --private-key=env:COSIGN_PRIVATE_KEY \
      --password=env:COSIGN_PASSWORD \
      --registry-username=$GITHUB_USERNAME \
      --registry-password=env:GITHUB_TOKEN



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
