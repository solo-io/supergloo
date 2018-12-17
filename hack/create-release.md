## Release process for supergloo

1. Delete `_output` folder: `rm -rf _output`
2. Create new release binaries: `make release-binaries`
3. Create release in github and upload files: `GITHUB_TOKEN=<token> VERSION=<version> make release`
    * token is generated from github and stored locally. Tutorial [here](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/)
    * version is chosen manually. It is incremented from the previous version. Current tag available here `git describe --tags`