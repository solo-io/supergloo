#!/bin/bash

set -ex

curl -sSL https://raw.githubusercontent.com/istio/api/7b94541b038b4dcc78e99a24262abeef20bf88af/rbac/v1alpha1/rbac.proto > rbac.proto

# add imports

sed -i -e 's$go_package="istio.io/api/rbac/v1alpha1"$go_package="github.com/solo-io/supergloo/pkg/api/external/istio/rbac/v1alpha1"$' rbac.proto
sed -i -e 's/istio.rbac.v1alpha1/rbac.istio.io' rbac.proto
sed -i -e "/option go_package/r imports.txt" rbac.proto
# inject fields to ServiceRole and ServiceRoleBinding
sed -i -e "/message ServiceRole/r fields.txt" rbac.proto
# inject fields to RbacConfig
sed -i -e "/message RbacConfig/r fields.txt" rbac.proto

sed -i -e "/message ServiceRole {/i//@solo-kit:resource.short_name=svcrole" rbac.proto
sed -i -e "/message ServiceRole {/i//@solo-kit:resource.plural_name=service_roles" rbac.proto
sed -i -e "/message ServiceRole {/i//@solo-kit:resource.resource_groups=rbac.istio.io" rbac.proto

sed -i -e "/message ServiceRoleBinding {/i//@solo-kit:resource.short_name=svcrolebind" rbac.proto
sed -i -e "/message ServiceRoleBinding {/i//@solo-kit:resource.plural_name=service_role_bindings" rbac.proto
sed -i -e "/message ServiceRoleBinding {/i//@solo-kit:resource.resource_groups=rbac.istio.io" rbac.proto

sed -i -e "/message RbacConfig {/i//@solo-kit:resource.short_name=rbaccfg" rbac.proto
sed -i -e "/message RbacConfig {/i//@solo-kit:resource.plural_name=rbac_configs" rbac.proto
sed -i -e "/message RbacConfig {/i//@solo-kit:resource.resource_groups=rbac.istio.io" rbac.proto