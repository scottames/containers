args := ""
progress := if args != "" { "auto" } else { "plain" }
labels := ""
tags := "test"

# renovate: datasource=docker depName=quay.io/fedora-ostree-desktops/silverblue
tagFedoraLatestVersion := "40"


_default:
  @just --list --list-heading $'' --list-prefix $''

# run `dagger develop` for all modules
develop:
    #!/usr/bin/env bash
    go work init
    for dir in */; do
      if [[ -f "${dir}/dagger.json" ]]; then
        dagger develop -m "${dir}"
        go work use "${dir}"
      fi
    done

# initialize a new Dagger module
[no-exit-message]
init module:
    @test ! -d {{module}} || (echo "Module \"{{module}}\" already exists" && exit 1)

    mkdir -p {{module}}
    cd {{module}} && dagger init --sdk go --name {{module}} --source .
    dagger develop -m {{module}}

[no-exit-message]
install target module:
  pushd {{ target }}
  dagger install ../{{ module }}
  popd

import 'atomic/justfile'
