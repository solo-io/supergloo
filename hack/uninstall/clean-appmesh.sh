#!/usr/bin/env bash
MESH_NAME=$(aws appmesh list-meshes | jq -r .meshes[0].meshName); for i in $(aws appmesh list-virtual-nodes --mesh-name $MESH_NAME | jq .virtualNodes[].virtualNodeName -r); do aws appmesh delete-virtual-node --mesh-name $MESH_NAME --virtual-node-name $i; done
MESH_NAME=$(aws appmesh list-meshes | jq -r .meshes[0].meshName); for i in $(aws appmesh list-virtual-routers --mesh-name $MESH_NAME | jq .virtualRouters[].virtualRouterName -r); do aws appmesh delete-virtual-router --mesh-name $MESH_NAME --virtual-router-name $i; done
aws appmesh delete-mesh --mesh-name $(aws appmesh list-meshes | jq -r .meshes[0].meshName)
