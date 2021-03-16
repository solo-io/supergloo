#!/usr/bin/env bash

###################################################################################
# This script generates a versioned docs website for Gloo Mesh and
# deploys to Firebase.
###################################################################################

set -ex

# Update this array with all versions of GlooMesh to include in the versioned docs website.
readarray -t versions < <(jq -r '.versions[]' version.json)
latestVersion=$(jq -r '.latest' version.json)

# Firebase configuration
firebaseJson=$(cat <<EOF
{
  "hosting": {
    "site": "gloo-mesh",
    "public": "public",
    "ignore": [
      "firebase.json",
      "**/.*",
      "**/node_modules/**"
    ],
    "rewrites": [
      {
        "source": "/",
        "destination": "/gloo-mesh/latest/index.html"
      },
      {
        "source": "/gloo-mesh",
        "destination": "/gloo-mesh/latest/index.html"
      }
    ]
  }
}
EOF
)

# This script assumes that the working directory is the root of the repo.
workingDir=$(pwd)
docsSiteDir=$workingDir/ci
repoDir=$workingDir/gloo-mesh-temp

mkdir -p "$docsSiteDir"
echo "$firebaseJson" > "$docsSiteDir/firebase.json"

git clone https://github.com/solo-io/gloo-mesh.git "$repoDir"

export PATH=$workingDir/_output/.bin:$PATH

# install go tools to sub-repo
make -C "$repoDir" install-go-tools

# Generates a data/Solo.yaml file with $1 being the specified version.
function generateHugoVersionsYaml() {
  local yamlFile=$repoDir/docs/data/Solo.yaml
  {
    echo "LatestVersion: $latestVersion"
    # /gloo-mesh prefix is needed because the site is hosted under a domain name with suffix /gloo-mesh
    echo "DocsVersion: /gloo-mesh/$1"
    echo "CodeVersion: $1"
    echo "DocsVersions:"
    for hugoVersion in "${versions[@]}"; do
      echo "  - $hugoVersion"
    done
  } > "$yamlFile"
}

for version in "${versions[@]}"; do
  echo "Generating site for version $version"
  cd "$repoDir"
  if [[ "$version" == "main" ]]; then
    git checkout main
  else
    git checkout tags/v"$version"
  fi

  # Replace version with "latest" if it's the latest version. This enables URLs with "/latest/..."
  [[ "$version" ==  "$latestVersion" ]] && version="latest"

  cd docs

  # Generate data/Solo.yaml file with version info populated.
  generateHugoVersionsYaml $version

  # Use partials from master
  mkdir -p layouts/partials
  cp -a "$workingDir/layouts/partials/." layouts/partials/
  cp -f "$workingDir/Makefile" Makefile
  cp -af "$workingDir/docsgen/." docsgen
  mkdir -p cmd
  cp -f "$workingDir/cmd/docsgen.go" cmd/docsgen.go
  # Generate the versioned static site.
  make site-release

  # Copy over versioned static site to firebase content folder.
  mkdir -p "$docsSiteDir/public/gloo-mesh/$version"
  cp -a site-latest/. "$docsSiteDir/public/gloo-mesh/$version/"

  # If we are on the latest version, then copy over `404.html` so firebase uses that.
  # https://firebase.google.com/docs/hosting/full-config#404
  [[ "$version" ==  "latest" ]] && cp site-latest/404.html "$docsSiteDir/public/404.html"

  # Discard git changes and vendor_any for subsequent checkouts
  cd "$repoDir"
  git reset --hard
  rm -fr vendor_any
done
