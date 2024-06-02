#!/bin/env bash

. /etc/os-release

echo "name - id - version: $NAME - $ID - $VERSION_ID" >>/usr/share/ublue-os/foo.txt
echo "pretty name: $PRETTY_NAME" >>/usr/share/ublue-os/foo.txt
