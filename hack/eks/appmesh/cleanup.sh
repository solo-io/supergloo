#!/usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# Read MESH name
MESH_NAME=$1
if [[ -z "${MESH_NAME}" ]];
then
    echo "Must provide mesh name"
    exit 1
fi

# Delete NODES
for node in $(aws appmesh list-virtual-nodes --mesh-name ${MESH_NAME} | jq .virtualNodes[].virtualNodeName -r); do
    aws appmesh delete-virtual-node --mesh-name ${MESH_NAME} --virtual-node-name ${node} >/dev/null
    echo "Deleted VIRTUAL NODE: ${node}"
done

# Delete VIRTUAL SERVICES
for vs in $(aws appmesh list-virtual-services --mesh-name ${MESH_NAME} | jq .virtualServices[].virtualServiceName -r); do
    aws appmesh delete-virtual-service --mesh-name ${MESH_NAME} --virtual-service-name ${vs} >/dev/null
    echo "Deleted VIRTUAL SERVICE: ${vs}"
done

# Delete VIRTUAL ROUTERS and ROUTES
for router in $(aws appmesh list-virtual-routers --mesh-name ${MESH_NAME} | jq .virtualRouters[].virtualRouterName -r); do
    # Delete each ROUTE linked to the VIRTUAL ROUTER
    for route in $(aws appmesh list-routes --mesh-name ${MESH_NAME} --virtual-router-name ${router} | jq .routes[].routeName -r); do
        aws appmesh delete-route --mesh-name ${MESH_NAME} --virtual-router-name ${router} --route-name ${route} >/dev/null
        echo "Deleted ROUTE: ${route}"
    done;

    aws appmesh delete-virtual-router --mesh-name ${MESH_NAME} --virtual-router-name ${router} >/dev/null
    echo "Deleted VIRTUAL ROUTER: ${router}"
done


aws appmesh delete-mesh --mesh-name ${MESH_NAME} >/dev/null
echo "Deleted MESH: ${MESH_NAME}"
