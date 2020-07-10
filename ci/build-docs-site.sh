#!/bin/bash

###################################################################################
# This script generates a versioned docs website for Service Mesh Hub and
# deploys to Firebase.
###################################################################################

set -ex

# Update this array with all versions of SMH to include in the versioned docs website.
declare -a versions=($(cat docs/version.json | jq -rc '."versions" | join(" ")'))
latestVersion=$(cat docs/version.json | jq -r ."latest")

# Firebase configuration
firebaseJson=$(cat <<EOF
{
  "hosting": {
    "site": "service-mesh-hub",
    "public": "public",
    "ignore": [
      "firebase.json",
      "**/.*",
      "**/node_modules/**"
    ],
    "rewrites": [
      {
        "source": "/",
        "destination": "/latest/index.html"
      }
    ]
  }
}
EOF
)

# This script assumes that the working directory is the root of the repo.
workingDir=$(pwd)
docsSiteDir=$workingDir/ci/docs-site
repoDir=$workingDir/ci/service-mesh-hub-temp

mkdir $docsSiteDir
echo $firebaseJson > $docsSiteDir/firebase.json

git clone https://github.com/solo-io/service-mesh-hub.git $repoDir

# Generates a data/Solo.yaml file with $1 being the specified version.
function generateHugoVersionsYaml() {
  yamlFile=$repoDir/docs/data/Solo.yaml
  # Truncate file first.
  echo "DocsVersion: /$1" > $yamlFile
  echo "CodeVersion: $1" >> $yamlFile
  echo "DocsVersions:" >> $yamlFile
  for hugoVersion in "${versions[@]}"
  do
    echo "  - $hugoVersion" >> $yamlFile
  done
}

for version in "${versions[@]}"
do
  echo "Generating site for version $version"
  cd $repoDir
  if [[ "$version" == "master" ]]
  then
    git checkout master
  else
    git checkout tags/v"$version"
  fi
  # Replace version with "latest" if it's the latest version. This enables URLs with "/latest/..."
  if [[ "$version" ==  "$latestVersion" ]]
  then
    version="latest"
  fi
  go run codegen/docs/docsgen.go
  cd docs
  # Generate data/Solo.yaml file with version info populated.
  generateHugoVersionsYaml $version
  # Generate the versioned static site.
  make site-release
  # Copy over versioned static site to firebase content folder.
  mkdir -p $docsSiteDir/public/$version
  cp -a site-latest/. $docsSiteDir/public/$version/
done

# Deploy the complete versioned site to Firebase. Firebase credentials are implicitly passed through the FIREBASE_TOKEN env var.
cd $docsSiteDir
firebase use --add solo-corp
firebase deploy --only hosting:service-mesh-hub

# Clean up directories.
rm -fr $docsSiteDir
rm -fr $repoDir
