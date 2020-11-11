#!/bin/sh

set -eu

if ! [ -x "$(command -v python)" ]; then
     echo Python is required to install meshctl
     exit 1
 fi

if [ -z "${GLOO_MESH_VERSION:-}" ]; then
  GLOO_MESH_VERSIONS=$(curl -sH"Accept: application/vnd.github.v3+json" https://api.github.com/repos/solo-io/gloo-mesh/releases | python -c "import sys; from distutils.version import LooseVersion; from json import loads as l; releases = l(sys.stdin.read()); releases = [release['tag_name'] for release in releases];  releases.sort(key=LooseVersion, reverse=True); print('\n'.join(releases))")
else
  GLOO_MESH_VERSIONS="${GLOO_MESH_VERSION}"
fi

if [ "$(uname -s)" = "Darwin" ]; then
  OS=darwin
else
  OS=linux
fi

for gloo_mesh_version in $GLOO_MESH_VERSIONS; do

tmp=$(mktemp -d /tmp/gloo_mesh.XXXXXX)
filename="meshctl-${OS}-amd64"
url="https://github.com/solo-io/gloo-mesh/releases/download/${gloo_mesh_version}/${filename}"

if curl -f ${url} >/dev/null 2>&1; then
  echo "Attempting to download meshctl version ${gloo_mesh_version}"
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
  mkdir -p ".gloo-mesh/bin"
  mv "${tmp}/${filename}" ".gloo-mesh/bin/meshctl"
  chmod +x ".gloo-mesh/bin/meshctl"
)

rm -r "$tmp"

echo "meshctl was successfully installed 🎉"
echo ""
echo "Add the Gloo Mesh CLI to your path with:"
echo "  export PATH=\$HOME/.gloo-mesh/bin:\$PATH"
echo ""
echo "Now run:"
echo "  meshctl install     # install Gloo Mesh management plane"
echo "Please see visit the Gloo Mesh website for more info:  https://www.solo.io/products/gloo-mesh/"
exit 0
done

echo "No versions of meshctl found."
exit 1
