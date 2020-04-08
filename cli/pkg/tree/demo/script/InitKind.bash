# Params:
# name of the cluster

name=%s

if [ -z "$name" ]; then
  name=cluster-$(xxd -l16 -ps /dev/urandom)
fi

kind create cluster --name="$name"
