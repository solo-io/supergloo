## Release process for supergloo

1. Delete `_output` folder: `rm -rf _output`
2. Create new release binaries: `make release-binaries`
3. Create release in github and upload files: `GITHUB_TOKEN=<token> VERSION=<version> make release`
