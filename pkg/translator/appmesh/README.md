# AWS App Mesh translator
This package contains the logic for translating a set of Supergloo Routing Rules to a set of AWS App Mesh resources.

Following are is a high-level description of the steps that are executed during translation for each mesh. The 
translation can be divided into three stages:

## Initialization
Here we parse the API snapshot for App Mesh related resources and initialize a set of **Virtual Node**s.

1. For each pod in the API Snapshot:
    1. check if the pod contains the Appmesh sidecar by looking for containers with the **APPMESH_VIRTUAL_NODE_NAME** env
    2. if it does:
        - determine if it belongs to the mesh we are translating
        - determine which virtual node it belongs to
        - check if it has an initContainer with the APPMESH_APP_PORTS env:
            - if yes and it's not empty, keep track of the ports
            - else fail
2. Group all the pods that belong to the mesh by the Virtual Node names they specify to
3. For each of the Virtual Node names:
    1. collect:
        - all the upstreams that match the pods that belong to the Virtual Node
        - all the ports for the above upstreams
        - all the ports specified via the **APPMESH_APP_PORTS** on each pod
    2. find **hostname**:
        - if no upstreams were found, fail
        - if the upstreams don't have the same host, fail, as we are not able to determine which hostname to use for the Virtual Node
        - if all upstreams have the same host, use the **hostname** as the DNS service discovery hostname for the Virtual Node
    3. find ports: verify that the set of upstream ports contains the set of **APPMESH_APP_PORTS**; if not, fail
    4. create a **Virtual Node** object with the following attributes:
        - VirtualNodeName: name of the VN as specified in the **APPMESH_VIRTUAL_NODE_NAME** env
        - Spec.ServiceDiscovery.Dns.Hostname: the **hostname** determined earlier
        - Spec.Listeners: one listener for each of the **APPMESH_APP_PORTS** 
        - Spec.Backends: nil, will be added later on
        
## Apply routing rules
At this point we have determined the set of Virtual Nodes and can start processing the Routing Rules.

1. For each traffic shifting Routing Rule (other rule types are currently not supported):
    1. for each matcher in the rule, create an **HttpRoute** with
        - Match: an HttpRouteMatch created from the matcher 
        - Action: an HttpRouteAction according to the rule spec (will be the same for all the routes):
            - when processing the rule spec, ensure that all weighted destinations have the same port
    2. collect the unique hosts for all upstreams that match the destination selector
    3. build a map that associates a destination host with the set of HttpRoutes (same for each hostname)
2. After all Routing Rules have been processed, merge the resulting route maps
3. For each entry in the map (host/set of HttpRoutes):
    1. verify that the routes are to the same port on the host, fail if not
    2. create a **Virtual Router** with:
        - VirtualRouterName: name of the host
        - Listeners: exactly one listener for the port the hosts are listening on
    3. for each HttpRoute in the route set, create a **Route** with:
        - RouteName: "hostname-i" (where i is an integer starting at 0)
        - VirtualRouterName: the name of the previously created VR
    4. create a **Virtual Service** with:
        - VirtualServiceName: the host name
        - Provider (type VirtualRouter): the name of the previously created Virtual Router
    5. add the Virtual Service as **Backend** for all the existing Virtual Nodes (except the one wth the same hostname)

## Allow all traffic
After all Routing Rules have been processed, we update the Virtual Nodes in order for each one of them to be able to 
connect to the other ones. This is equivalent to applying a routing rule which allows all traffic from all other hosts 
for each host in the mesh.

1. For each **Virtual Node**:
    - If there does not exists a **Virtual Service** with the same hostname, create it with:
        - VirtualServiceName: the host name
        - Provider (type VirtualNode): the name of the **Virtual Node**
    - This Virtual Service will simply forward all traffic matching the hostname to the Virtual Node
2. For each **Virtual Node**:
    - add each existing Virtual Service as a **Backend**:
      - if it does not have the same host name
      - if the Virtual Node already has a Backend for this host name (created as part of a Routing Rule)
 


    
    
    
    
  