args := ""
gitRoot := `git rev-parse --show-toplevel`
goUpdates :="false"
labels := ""
progress := if args != "" { "auto" } else { "plain" }
tags := ""

# renovate: datasource=docker depName=quay.io/fedora-ostree-desktops/silverblue
tagFedoraLatestVersion := "43"


_default:
  @just --list --list-heading $'' --list-prefix $''

# run go updates for the given project (USE WITH CAUTION)
go-update project version="latest":
    #!/usr/bin/env bash
    echo "=> go update: {{ project }}"
    pushd "{{ project }}" >/dev/null || exit 1
    if [[ ! -f "go.mod" ]]; then
      echo "‼️ ERROR: no go.mod in {{ project }}"
      exit 1
    fi
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
    set -e
    _DAGGER_MODS=()
    if [[ -n "{{ mod }}" ]]; then
      _DAGGER_MODS=("{{ mod }}")
    else
      shopt -s globstar nullglob
      for _DAGGER_JSON in **/dagger.json; do
        _DAGGER_MODS+=("$(dirname "${_DAGGER_JSON}")")
      done
    fi

    for _DAGGER_MOD in "${_DAGGER_MODS[@]}"; do
      echo "=> ${_DAGGER_MOD}: dagger develop"

      pushd "${_DAGGER_MOD}" >/dev/null || exit
      _DAGGER_MOD_SOURCE="$(dagger config --silent --json | jq -r '.source')"

      # NOTE: use with caution!
      # Dagger is opinionated about the go version compatibility. It will barf
      # if the go version is greater than supported
      if [[ "{{ goUpdates }}" = "true" ]]; then
        _DAGGER_GO_MOD="${_DAGGER_MOD}/${_DAGGER_MOD_SOURCE}"
        echo "=> ${_DAGGER_GO_MOD}: go update"
        just -f "{{ gitRoot }}/justfile" go-update "${_DAGGER_GO_MOD}"
      fi

      dagger develop

      # remove generated bits we don't want
      rm -f LICENSE

      just -f "{{ gitRoot }}/justfile" go-work "${_DAGGER_MOD}"

      popd >/dev/null || exit 1
    done
    echo "=> dagger-develop: done"

# run `dagger update` for all Dagger modules, or the given module
update-dagger-dependencies mod="":
    #!/usr/bin/env bash
    set -euo pipefail
    _DAGGER_MODS=()
    if [[ -n "{{ mod }}" ]]; then
      _DAGGER_MODS=("{{ mod }}")
    else
      shopt -s globstar nullglob
      for _DAGGER_JSON in **/dagger.json; do
        _DAGGER_MODS+=("$(dirname "${_DAGGER_JSON}")")
      done
    fi

    for _DAGGER_MOD in "${_DAGGER_MODS[@]}"; do
      echo "=> ${_DAGGER_MOD}: dagger update"

      _DAGGER_CONFIG_JSON="$(dagger config --mod "${_DAGGER_MOD}" --silent --json)"
      _DAGGER_DEPS_OUTPUT="$(jq -r '.dependencies // [] | .[].source' <<<"${_DAGGER_CONFIG_JSON}")"
      _DAGGER_DEPS=()

      if [[ -n "${_DAGGER_DEPS_OUTPUT}" ]]; then
        mapfile -t _DAGGER_DEPS <<<"${_DAGGER_DEPS_OUTPUT}"
      fi

      if [[ "${#_DAGGER_DEPS[@]}" -eq 0 ]]; then
        continue
      fi

      dagger update --mod "${_DAGGER_MOD}" "${_DAGGER_DEPS[@]}"
    done

# initialize a new Dagger module
[no-exit-message]
init module:
  #!/usr/bin/env bash
  set -euo pipefail
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

# update scottames/daggerverse deps; use vX.Y.Z for all modules or <module>/vX.Y.Z for one module
update-scottames-daggerverse version mod="":
  #!/usr/bin/env bash
  set -euo pipefail
  _DAGGER_MODS=()
  _REQUESTED_MOD="{{ mod }}"
  _TARGET_MODULE=""
  _TARGET_VERSION="{{ version }}"
  if [[ "${_TARGET_VERSION}" == */* ]]; then
    _TARGET_MODULE="${_TARGET_VERSION%%/*}"
    _TARGET_VERSION="${_TARGET_VERSION#*/}"
  fi

  if [[ -n "${_REQUESTED_MOD}" ]]; then
    _DAGGER_MODS=("${_REQUESTED_MOD}")
  else
    mapfile -t _DAGGER_MODS < <(find . -type f -name dagger.json -print0 | xargs -0 dirname)
  fi

  for _DAGGER_MOD in "${_DAGGER_MODS[@]}"; do
    echo "=> ${_DAGGER_MOD} @ {{ version }}"
    pushd "${_DAGGER_MOD}"
    _DAGGER_CONFIG_JSON="$(dagger config --silent --json)"
    mapfile -t _DAGGER_DEPS < <(
      jq -r '.dependencies // [] | .[].source | select(contains("scottames/daggerverse")) | split("@")[0]' \
        <<<"${_DAGGER_CONFIG_JSON}"
    )

    for mod_to_update in "${_DAGGER_DEPS[@]}"
    do
      _DAGGERVERSE_MODULE="${mod_to_update##*/}"
      if [[ -n "${_TARGET_MODULE}" && "${_DAGGERVERSE_MODULE}" != "${_TARGET_MODULE}" ]]; then
        continue
      fi

      echo "=> update: ${mod_to_update}@${_DAGGERVERSE_MODULE}/${_TARGET_VERSION}"
      dagger install "${mod_to_update}@${_DAGGERVERSE_MODULE}/${_TARGET_VERSION}"
    done
    popd
  done

import 'atomic/justfile'
import 'toolbox/fedora/justfile'
