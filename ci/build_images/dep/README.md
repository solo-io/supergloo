This code is based from code was taken from here:
https://github.com/GoogleCloudPlatform/cloud-builders-community/pull/69/files/cf76def06fc9cc6a99c0e32c088492023b8378ee

I used the short version to build it:
```
docker build -t soloio/dep .
docker push soloio/dep
```

-----------------

# [go dep](https://github.com/golang/dep)
The Dockerfile and scripts here help you use Google Cloud Builder to launch the **go dep** tool.
## Building this builder
To build this builder, run the following command in this directory.
    $ gcloud container builds submit . --config=cloudbuild.yaml
## Example
```yaml
steps:
# Make sure all dependencies are in the desired state
- name: 'gcr.io/$PROJECT_ID/dep'
  args: ['ensure', '-v']
  env: ['PROJECT_ROOT=github.com/myorg/myproject']
  id: 'dep'
```