#!/bin/sh

set -eu

if ! [ -x "$(command -v python)" ]; then
     echo Python is required to install meshctl
     exit 1
 fi

if [ -z "${SMH_VERSION:-}" ]; then
  SMH_VERSIONS=$(curl -sH"Accept: application/vnd.github.v3+json" https://api.github.com/repos/solo-io/service-mesh-hub/releases | python -c "import sys; from distutils.version import LooseVersion; from json import loads as l; releases = l(sys.stdin.read()); releases = [release['tag_name'] for release in releases];  releases.sort(key=LooseVersion, reverse=True); print('\n'.join(releases))")
else
  SMH_VERSIONS="${SMH_VERSION}"
fi

if [ "$(uname -s)" = "Darwin" ]; then
  OS=darwin
else
  OS=linux
fi

for smh_version in $SMH_VERSIONS; do

tmp=$(mktemp -d /tmp/smh.XXXXXX)
filename="meshctl-${OS}-amd64"
url="https://github.com/solo-io/service-mesh-hub/releases/download/${smh_version}/${filename}"

if curl -f ${url} >/dev/null 2>&1; then
  echo "Attempting to download meshctl version ${smh_version}"
else
  continue
fi

(
  cd "$tmp"

  echo "Downloading ${filename}..."

  SHA=$(curl -sL "${url}.sha256" | cut -d' ' -f1)
  curl -sLO "${url}"
  echo "Download complete!, validating checksum..."
  checksum=$(openssl dgst -sha256 "${filename}" | awk '{ print $2 }')
  if [ "$checksum" != "$SHA" ]; then
    echo "Checksum validation failed." >&2
    exit 1
  fi
  echo "Checksum valid."
)

(
  cd "$HOME"
  mkdir -p ".service-mesh-hub/bin"
  mv "${tmp}/${filename}" ".service-mesh-hub/bin/meshctl"
  chmod +x ".service-mesh-hub/bin/meshctl"
)

rm -r "$tmp"

echo "meshctl was successfully installed 🎉"
echo ""
echo "Add the Service Mesh Hub CLI to your path with:"
echo "  export PATH=\$HOME/.service-mesh-hub/bin:\$PATH"
echo ""
echo "Now run:"
echo "  meshctl install     # install Service Mesh Hub management plane"
echo "Please see visit the Service Mesh Hub website for more info:  https://www.solo.io/products/service-mesh-hub/"
exit 0
done

echo "No versions of meshctl found."
exit 1