args := ""
gitRoot := `git rev-parse --show-toplevel`
goUpdates :="false"
labels := ""
progress := if args != "" { "auto" } else { "plain" }
tags := ""

# renovate: datasource=docker depName=quay.io/fedora-ostree-desktops/silverblue
tagFedoraLatestVersion := "40"


_default:
  @just --list --list-heading $'' --list-prefix $''

# run go updates for the given project (USE WITH CAUTION)
go-update project version="latest":
    #!/usr/bin/env bash
    pushd "{{ project }}" >/dev/null || exit 1
    [ -x "$(command -v gobrew)" ] || exit 1
    gobrew use "{{ version }}"
    # remove the go version, let the mod update it
    sed -i '/^go\s.*$/d' go.mod
    go get -u
    go mod tidy
    popd >/dev/null || exit 1

# init go.work | https://go.dev/doc/tutorial/workspaces
go-work target="":
    #!/usr/bin/env bash

    pushd {{ gitRoot }} >/dev/null

    if [[ ! -f "go.work" ]]; then # only create go.work if not exists
      echo "=> go work init"
      go work init
    fi

    if [[ -n "{{ target }}" ]]; then # generate just for the given target
      echo "=> use: {{ target }}"
      go work use {{ target }}

    else # generate go.work with all dirs containing go.mod
      for _GO_MOD_DIR in $(find . -type f -name go.mod | xargs dirname); do
        echo "=> use: ${_GO_MOD_DIR}"
        go work use "${_GO_MOD_DIR}"
      done
    fi

# run `dagger develop` for all Dagger modules, or the given module
develop mod="":
    #!/usr/bin/env bash
    _DAGGER_MODS="{{ mod }}"
    if [[ -z "${_DAGGER_MODS}" ]]; then
      mapfile -t _DAGGER_MODS < <(find . -type f -name dagger.json -print0 | xargs -0 dirname)
    fi

    for _DAGGER_MOD in "${_DAGGER_MODS[@]}"; do
      pushd "${_DAGGER_MOD}" >/dev/null || exit
      _DAGGER_MOD_SOURCE="$(dagger config --silent --json | jq -r '.source')"

      # NOTE: use with caution!
      # Dagger is opinionated about the go version compatibility. It will barf
      # if the go version is greater than supported
      if [[ "{{ goUpdates }}" = "true" ]]; then
        echo "=> ${_DAGGER_MOD}: go update"
        just -f "{{ gitRoot }}/justfile" go-update "${_DAGGER_MOD}"
      fi

      echo "=> ${_DAGGER_MOD}: dagger develop"
      dagger develop

      # remove generated bits we don't want
      rm -f LICENSE

      just -f "{{ gitRoot }}/justfile" go-work "${_DAGGER_MOD}"

      popd >/dev/null || exit 1
    done

# initialize a new Dagger module
[no-exit-message]
init module:
  #!/usr/bin/env bash
  set -euxo pipefail
  test ! -d {{module}} \
  || (echo "Module \"{{module}}\" already exists" && exit 1)

  mkdir -p {{module}}
  cd {{module}} && dagger init --sdk go --name {{module}} --source .
  dagger develop -m {{module}}

[no-exit-message]
install target module:
  pushd {{ target }}
  dagger install {{ module }}
  popd

update-scottames-daggerverse version mod="":
  #!/usr/bin/env bash
  _DAGGER_MODS="{{ mod }}"
  if [[ -z "${_DAGGER_MODS}" ]]; then
    mapfile -t _DAGGER_MODS < <(find . -type f -name dagger.json -print0 | xargs -0 dirname)
  fi

  for _DAGGER_MOD in "${_DAGGER_MODS[@]}"; do
    echo "=> ${_DAGGER_MOD}"
    pushd "${_DAGGER_MOD}"
    for mod_to_update in $( dagger config --silent --json \
      | jq -r '.dependencies | .[].source' | grep 'scottames/daggerverse' \
      | cut -d'@' -f1)
    do
      echo "=> update: ${mod_to_update}"
      dagger install "${mod_to_update}@{{ version }}"
    done
    popd
  done

import 'atomic/justfile'
import 'toolbox/fedora/justfile'
