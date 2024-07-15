args := ""
progress := if args != "" { "auto" } else { "plain" }
labels := ""
tags := "test"

# renovate: datasource=docker depName=quay.io/fedora-ostree-desktops/silverblue
tagFedoraLatestVersion := "40"


_default:
  @just --list --list-heading $'' --list-prefix $''

# run `dagger develop` for all modules (set update=true to run go updates)
develop update="false":
    #!/usr/bin/env bash
    test -f go.work || go work init
    for _DAGGER_MOD in $(find . -type f -name dagger.json | xargs dirname); do
      echo "=> $_DAGGER_MOD"
      pushd $_DAGGER_MOD > /dev/null
      gobrew use mod
      dagger develop
      if [[ "{{ update }}" = "true" ]]; then
        go get -u && go mod tidy
      fi
      rm -f LICENSE # remove generated LICENSE
      popd > /dev/null
      go work use "${_DAGGER_MOD}"
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
import 'toolbox/fedora/justfile'
