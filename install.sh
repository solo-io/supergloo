#! /bin/sh

set -eu


SUPERGLOO_VERSIONS=$(curl -sH"Accept: application/vnd.github.v3+json" https://api.github.com/repos/solo-io/supergloo/releases | python -c "import sys; from json import loads as l; releases = l(sys.stdin.read()); print('\n'.join(release['tag_name'] for release in releases))")

if [ "$(uname -s)" = "Darwin" ]; then
  OS=darwin
else
  OS=linux
fi

for SUPERGLOO_VERSION in $SUPERGLOO_VERSIONS; do

tmp=$(mktemp -d /tmp/supergloo.XXXXXX)
filename="supergloo-cli-${OS}-amd64"
url="https://github.com/solo-io/supergloo/releases/download/${SUPERGLOO_VERSION}/${filename}"

if curl -f ${url} >/dev/null 2>&1; then
  echo "Attempting to download supergloo version ${SUPERGLOO_VERSIONS}"
else
  continue
fi

(
  cd "$tmp"

  echo "Downloading from ${url}..."

  SHA=$(curl -sL "${url}.sha256" | awk '{ print $1 }')
  curl -LO "${url}"
  echo ""
  echo "Download complete!, validating checksum..."
  checksum=$(openssl dgst -sha256 "${filename}" | awk '{ print $2 }')
  if [ "$checksum" != "$SHA" ]; then
    echo "Checksum validation failed." >&2
    exit 1
  fi
  echo "Checksum valid."
  echo ""
)

(
  cd "$HOME"
  mkdir -p ".supergloo/bin"
  mv "${tmp}/${filename}" ".supergloo/bin/supergloo"
  chmod +x ".supergloo/bin/supergloo"
)

rm -r "$tmp"

echo "SuperGloo was successfully installed 🎉"
echo ""
echo "Add the supergloo CLI to your path with:"
echo ""
echo "  export PATH=\$PATH:\$HOME/.supergloo/bin"
echo ""
echo "Now run:"
echo ""
echo "  supergloo init        # install supergloo into the 'supergloo-system' namespace"
echo ""
echo "Looking for more? Visit https://supergloo.solo.io/installation/"
echo ""
exit 0
done

echo "No versions of supergloo found."
exit 1
