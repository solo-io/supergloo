// Code generated by skv2. DO NOT EDIT.

//go:generate mockgen -source ./sets.go -destination mocks/sets.go

package v1alpha2sets



import (
    networking_mesh_gloo_solo_io_v1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"

    "github.com/rotisserie/eris"
    sksets "github.com/solo-io/skv2/contrib/pkg/sets"
    "github.com/solo-io/skv2/pkg/ezkube"
    "k8s.io/apimachinery/pkg/util/sets"
)

type TrafficPolicySet interface {
	// Get the set stored keys
    Keys() sets.String
    // List of resources stored in the set. Pass an optional filter function to filter on the list.
    List(filterResource ... func(*networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy) bool) []*networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy
    // Return the Set as a map of key to resource.
    Map() map[string]*networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy
    // Insert a resource into the set.
    Insert(trafficPolicy ...*networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy)
    // Compare the equality of the keys in two sets (not the resources themselves)
    Equal(trafficPolicySet TrafficPolicySet) bool
    // Check if the set contains a key matching the resource (not the resource itself)
    Has(trafficPolicy ezkube.ResourceId) bool
    // Delete the key matching the resource
    Delete(trafficPolicy  ezkube.ResourceId)
    // Return the union with the provided set
    Union(set TrafficPolicySet) TrafficPolicySet
    // Return the difference with the provided set
    Difference(set TrafficPolicySet) TrafficPolicySet
    // Return the intersection with the provided set
    Intersection(set TrafficPolicySet) TrafficPolicySet
    // Find the resource with the given ID
    Find(id ezkube.ResourceId) (*networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy, error)
    // Get the length of the set
    Length() int
}

func makeGenericTrafficPolicySet(trafficPolicyList []*networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy) sksets.ResourceSet {
    var genericResources []ezkube.ResourceId
    for _, obj := range trafficPolicyList {
        genericResources = append(genericResources, obj)
    }
    return sksets.NewResourceSet(genericResources...)
}

type trafficPolicySet struct {
    set sksets.ResourceSet
}

func NewTrafficPolicySet(trafficPolicyList ...*networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy) TrafficPolicySet {
    return &trafficPolicySet{set: makeGenericTrafficPolicySet(trafficPolicyList)}
}

func NewTrafficPolicySetFromList(trafficPolicyList *networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicyList) TrafficPolicySet {
    list := make([]*networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy, 0, len(trafficPolicyList.Items))
    for idx := range trafficPolicyList.Items {
        list = append(list, &trafficPolicyList.Items[idx])
    }
    return &trafficPolicySet{set: makeGenericTrafficPolicySet(list)}
}

func (s *trafficPolicySet) Keys() sets.String {
	if s == nil {
		return sets.String{}
    }
    return s.set.Keys()
}

func (s *trafficPolicySet) List(filterResource ... func(*networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy) bool) []*networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy {
    if s == nil {
        return nil
    }
    var genericFilters []func(ezkube.ResourceId) bool
    for _, filter := range filterResource {
        genericFilters = append(genericFilters, func(obj ezkube.ResourceId) bool {
            return filter(obj.(*networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy))
        })
    }

    var trafficPolicyList []*networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy
    for _, obj := range s.set.List(genericFilters...) {
        trafficPolicyList = append(trafficPolicyList, obj.(*networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy))
    }
    return trafficPolicyList
}

func (s *trafficPolicySet) Map() map[string]*networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy {
    if s == nil {
        return nil
    }

    newMap := map[string]*networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy{}
    for k, v := range s.set.Map() {
        newMap[k] = v.(*networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy)
    }
    return newMap
}

func (s *trafficPolicySet) Insert(
        trafficPolicyList ...*networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy,
) {
    if s == nil {
        panic("cannot insert into nil set")
    }

    for _, obj := range trafficPolicyList {
        s.set.Insert(obj)
    }
}

func (s *trafficPolicySet) Has(trafficPolicy ezkube.ResourceId) bool {
    if s == nil {
        return false
    }
    return s.set.Has(trafficPolicy)
}

func (s *trafficPolicySet) Equal(
        trafficPolicySet TrafficPolicySet,
) bool {
    if s == nil {
        return trafficPolicySet == nil
    }
    return s.set.Equal(makeGenericTrafficPolicySet(trafficPolicySet.List()))
}

func (s *trafficPolicySet) Delete(TrafficPolicy ezkube.ResourceId) {
    if s == nil {
        return
    }
    s.set.Delete(TrafficPolicy)
}

func (s *trafficPolicySet) Union(set TrafficPolicySet) TrafficPolicySet {
    if s == nil {
        return set
    }
    return NewTrafficPolicySet(append(s.List(), set.List()...)...)
}

func (s *trafficPolicySet) Difference(set TrafficPolicySet) TrafficPolicySet {
    if s == nil {
        return set
    }
    newSet := s.set.Difference(makeGenericTrafficPolicySet(set.List()))
    return &trafficPolicySet{set: newSet}
}

func (s *trafficPolicySet) Intersection(set TrafficPolicySet) TrafficPolicySet {
    if s == nil {
        return nil
    }
    newSet := s.set.Intersection(makeGenericTrafficPolicySet(set.List()))
    var trafficPolicyList []*networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy
    for _, obj := range newSet.List() {
        trafficPolicyList = append(trafficPolicyList, obj.(*networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy))
    }
    return NewTrafficPolicySet(trafficPolicyList...)
}


func (s *trafficPolicySet) Find(id ezkube.ResourceId) (*networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy, error) {
    if s == nil {
        return nil, eris.Errorf("empty set, cannot find TrafficPolicy %v", sksets.Key(id))
    }
	obj, err := s.set.Find(&networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy{}, id)
	if err != nil {
		return nil, err
    }

    return obj.(*networking_mesh_gloo_solo_io_v1alpha2.TrafficPolicy), nil
}

func (s *trafficPolicySet) Length() int {
    if s == nil {
        return 0
    }
    return s.set.Length()
}

type AccessPolicySet interface {
	// Get the set stored keys
    Keys() sets.String
    // List of resources stored in the set. Pass an optional filter function to filter on the list.
    List(filterResource ... func(*networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy) bool) []*networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy
    // Return the Set as a map of key to resource.
    Map() map[string]*networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy
    // Insert a resource into the set.
    Insert(accessPolicy ...*networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy)
    // Compare the equality of the keys in two sets (not the resources themselves)
    Equal(accessPolicySet AccessPolicySet) bool
    // Check if the set contains a key matching the resource (not the resource itself)
    Has(accessPolicy ezkube.ResourceId) bool
    // Delete the key matching the resource
    Delete(accessPolicy  ezkube.ResourceId)
    // Return the union with the provided set
    Union(set AccessPolicySet) AccessPolicySet
    // Return the difference with the provided set
    Difference(set AccessPolicySet) AccessPolicySet
    // Return the intersection with the provided set
    Intersection(set AccessPolicySet) AccessPolicySet
    // Find the resource with the given ID
    Find(id ezkube.ResourceId) (*networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy, error)
    // Get the length of the set
    Length() int
}

func makeGenericAccessPolicySet(accessPolicyList []*networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy) sksets.ResourceSet {
    var genericResources []ezkube.ResourceId
    for _, obj := range accessPolicyList {
        genericResources = append(genericResources, obj)
    }
    return sksets.NewResourceSet(genericResources...)
}

type accessPolicySet struct {
    set sksets.ResourceSet
}

func NewAccessPolicySet(accessPolicyList ...*networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy) AccessPolicySet {
    return &accessPolicySet{set: makeGenericAccessPolicySet(accessPolicyList)}
}

func NewAccessPolicySetFromList(accessPolicyList *networking_mesh_gloo_solo_io_v1alpha2.AccessPolicyList) AccessPolicySet {
    list := make([]*networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy, 0, len(accessPolicyList.Items))
    for idx := range accessPolicyList.Items {
        list = append(list, &accessPolicyList.Items[idx])
    }
    return &accessPolicySet{set: makeGenericAccessPolicySet(list)}
}

func (s *accessPolicySet) Keys() sets.String {
	if s == nil {
		return sets.String{}
    }
    return s.set.Keys()
}

func (s *accessPolicySet) List(filterResource ... func(*networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy) bool) []*networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy {
    if s == nil {
        return nil
    }
    var genericFilters []func(ezkube.ResourceId) bool
    for _, filter := range filterResource {
        genericFilters = append(genericFilters, func(obj ezkube.ResourceId) bool {
            return filter(obj.(*networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy))
        })
    }

    var accessPolicyList []*networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy
    for _, obj := range s.set.List(genericFilters...) {
        accessPolicyList = append(accessPolicyList, obj.(*networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy))
    }
    return accessPolicyList
}

func (s *accessPolicySet) Map() map[string]*networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy {
    if s == nil {
        return nil
    }

    newMap := map[string]*networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy{}
    for k, v := range s.set.Map() {
        newMap[k] = v.(*networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy)
    }
    return newMap
}

func (s *accessPolicySet) Insert(
        accessPolicyList ...*networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy,
) {
    if s == nil {
        panic("cannot insert into nil set")
    }

    for _, obj := range accessPolicyList {
        s.set.Insert(obj)
    }
}

func (s *accessPolicySet) Has(accessPolicy ezkube.ResourceId) bool {
    if s == nil {
        return false
    }
    return s.set.Has(accessPolicy)
}

func (s *accessPolicySet) Equal(
        accessPolicySet AccessPolicySet,
) bool {
    if s == nil {
        return accessPolicySet == nil
    }
    return s.set.Equal(makeGenericAccessPolicySet(accessPolicySet.List()))
}

func (s *accessPolicySet) Delete(AccessPolicy ezkube.ResourceId) {
    if s == nil {
        return
    }
    s.set.Delete(AccessPolicy)
}

func (s *accessPolicySet) Union(set AccessPolicySet) AccessPolicySet {
    if s == nil {
        return set
    }
    return NewAccessPolicySet(append(s.List(), set.List()...)...)
}

func (s *accessPolicySet) Difference(set AccessPolicySet) AccessPolicySet {
    if s == nil {
        return set
    }
    newSet := s.set.Difference(makeGenericAccessPolicySet(set.List()))
    return &accessPolicySet{set: newSet}
}

func (s *accessPolicySet) Intersection(set AccessPolicySet) AccessPolicySet {
    if s == nil {
        return nil
    }
    newSet := s.set.Intersection(makeGenericAccessPolicySet(set.List()))
    var accessPolicyList []*networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy
    for _, obj := range newSet.List() {
        accessPolicyList = append(accessPolicyList, obj.(*networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy))
    }
    return NewAccessPolicySet(accessPolicyList...)
}


func (s *accessPolicySet) Find(id ezkube.ResourceId) (*networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy, error) {
    if s == nil {
        return nil, eris.Errorf("empty set, cannot find AccessPolicy %v", sksets.Key(id))
    }
	obj, err := s.set.Find(&networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy{}, id)
	if err != nil {
		return nil, err
    }

    return obj.(*networking_mesh_gloo_solo_io_v1alpha2.AccessPolicy), nil
}

func (s *accessPolicySet) Length() int {
    if s == nil {
        return 0
    }
    return s.set.Length()
}

type VirtualMeshSet interface {
	// Get the set stored keys
    Keys() sets.String
    // List of resources stored in the set. Pass an optional filter function to filter on the list.
    List(filterResource ... func(*networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh) bool) []*networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh
    // Return the Set as a map of key to resource.
    Map() map[string]*networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh
    // Insert a resource into the set.
    Insert(virtualMesh ...*networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh)
    // Compare the equality of the keys in two sets (not the resources themselves)
    Equal(virtualMeshSet VirtualMeshSet) bool
    // Check if the set contains a key matching the resource (not the resource itself)
    Has(virtualMesh ezkube.ResourceId) bool
    // Delete the key matching the resource
    Delete(virtualMesh  ezkube.ResourceId)
    // Return the union with the provided set
    Union(set VirtualMeshSet) VirtualMeshSet
    // Return the difference with the provided set
    Difference(set VirtualMeshSet) VirtualMeshSet
    // Return the intersection with the provided set
    Intersection(set VirtualMeshSet) VirtualMeshSet
    // Find the resource with the given ID
    Find(id ezkube.ResourceId) (*networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh, error)
    // Get the length of the set
    Length() int
}

func makeGenericVirtualMeshSet(virtualMeshList []*networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh) sksets.ResourceSet {
    var genericResources []ezkube.ResourceId
    for _, obj := range virtualMeshList {
        genericResources = append(genericResources, obj)
    }
    return sksets.NewResourceSet(genericResources...)
}

type virtualMeshSet struct {
    set sksets.ResourceSet
}

func NewVirtualMeshSet(virtualMeshList ...*networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh) VirtualMeshSet {
    return &virtualMeshSet{set: makeGenericVirtualMeshSet(virtualMeshList)}
}

func NewVirtualMeshSetFromList(virtualMeshList *networking_mesh_gloo_solo_io_v1alpha2.VirtualMeshList) VirtualMeshSet {
    list := make([]*networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh, 0, len(virtualMeshList.Items))
    for idx := range virtualMeshList.Items {
        list = append(list, &virtualMeshList.Items[idx])
    }
    return &virtualMeshSet{set: makeGenericVirtualMeshSet(list)}
}

func (s *virtualMeshSet) Keys() sets.String {
	if s == nil {
		return sets.String{}
    }
    return s.set.Keys()
}

func (s *virtualMeshSet) List(filterResource ... func(*networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh) bool) []*networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh {
    if s == nil {
        return nil
    }
    var genericFilters []func(ezkube.ResourceId) bool
    for _, filter := range filterResource {
        genericFilters = append(genericFilters, func(obj ezkube.ResourceId) bool {
            return filter(obj.(*networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh))
        })
    }

    var virtualMeshList []*networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh
    for _, obj := range s.set.List(genericFilters...) {
        virtualMeshList = append(virtualMeshList, obj.(*networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh))
    }
    return virtualMeshList
}

func (s *virtualMeshSet) Map() map[string]*networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh {
    if s == nil {
        return nil
    }

    newMap := map[string]*networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh{}
    for k, v := range s.set.Map() {
        newMap[k] = v.(*networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh)
    }
    return newMap
}

func (s *virtualMeshSet) Insert(
        virtualMeshList ...*networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh,
) {
    if s == nil {
        panic("cannot insert into nil set")
    }

    for _, obj := range virtualMeshList {
        s.set.Insert(obj)
    }
}

func (s *virtualMeshSet) Has(virtualMesh ezkube.ResourceId) bool {
    if s == nil {
        return false
    }
    return s.set.Has(virtualMesh)
}

func (s *virtualMeshSet) Equal(
        virtualMeshSet VirtualMeshSet,
) bool {
    if s == nil {
        return virtualMeshSet == nil
    }
    return s.set.Equal(makeGenericVirtualMeshSet(virtualMeshSet.List()))
}

func (s *virtualMeshSet) Delete(VirtualMesh ezkube.ResourceId) {
    if s == nil {
        return
    }
    s.set.Delete(VirtualMesh)
}

func (s *virtualMeshSet) Union(set VirtualMeshSet) VirtualMeshSet {
    if s == nil {
        return set
    }
    return NewVirtualMeshSet(append(s.List(), set.List()...)...)
}

func (s *virtualMeshSet) Difference(set VirtualMeshSet) VirtualMeshSet {
    if s == nil {
        return set
    }
    newSet := s.set.Difference(makeGenericVirtualMeshSet(set.List()))
    return &virtualMeshSet{set: newSet}
}

func (s *virtualMeshSet) Intersection(set VirtualMeshSet) VirtualMeshSet {
    if s == nil {
        return nil
    }
    newSet := s.set.Intersection(makeGenericVirtualMeshSet(set.List()))
    var virtualMeshList []*networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh
    for _, obj := range newSet.List() {
        virtualMeshList = append(virtualMeshList, obj.(*networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh))
    }
    return NewVirtualMeshSet(virtualMeshList...)
}


func (s *virtualMeshSet) Find(id ezkube.ResourceId) (*networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh, error) {
    if s == nil {
        return nil, eris.Errorf("empty set, cannot find VirtualMesh %v", sksets.Key(id))
    }
	obj, err := s.set.Find(&networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh{}, id)
	if err != nil {
		return nil, err
    }

    return obj.(*networking_mesh_gloo_solo_io_v1alpha2.VirtualMesh), nil
}

func (s *virtualMeshSet) Length() int {
    if s == nil {
        return 0
    }
    return s.set.Length()
}

type FailoverServiceSet interface {
	// Get the set stored keys
    Keys() sets.String
    // List of resources stored in the set. Pass an optional filter function to filter on the list.
    List(filterResource ... func(*networking_mesh_gloo_solo_io_v1alpha2.FailoverService) bool) []*networking_mesh_gloo_solo_io_v1alpha2.FailoverService
    // Return the Set as a map of key to resource.
    Map() map[string]*networking_mesh_gloo_solo_io_v1alpha2.FailoverService
    // Insert a resource into the set.
    Insert(failoverService ...*networking_mesh_gloo_solo_io_v1alpha2.FailoverService)
    // Compare the equality of the keys in two sets (not the resources themselves)
    Equal(failoverServiceSet FailoverServiceSet) bool
    // Check if the set contains a key matching the resource (not the resource itself)
    Has(failoverService ezkube.ResourceId) bool
    // Delete the key matching the resource
    Delete(failoverService  ezkube.ResourceId)
    // Return the union with the provided set
    Union(set FailoverServiceSet) FailoverServiceSet
    // Return the difference with the provided set
    Difference(set FailoverServiceSet) FailoverServiceSet
    // Return the intersection with the provided set
    Intersection(set FailoverServiceSet) FailoverServiceSet
    // Find the resource with the given ID
    Find(id ezkube.ResourceId) (*networking_mesh_gloo_solo_io_v1alpha2.FailoverService, error)
    // Get the length of the set
    Length() int
}

func makeGenericFailoverServiceSet(failoverServiceList []*networking_mesh_gloo_solo_io_v1alpha2.FailoverService) sksets.ResourceSet {
    var genericResources []ezkube.ResourceId
    for _, obj := range failoverServiceList {
        genericResources = append(genericResources, obj)
    }
    return sksets.NewResourceSet(genericResources...)
}

type failoverServiceSet struct {
    set sksets.ResourceSet
}

func NewFailoverServiceSet(failoverServiceList ...*networking_mesh_gloo_solo_io_v1alpha2.FailoverService) FailoverServiceSet {
    return &failoverServiceSet{set: makeGenericFailoverServiceSet(failoverServiceList)}
}

func NewFailoverServiceSetFromList(failoverServiceList *networking_mesh_gloo_solo_io_v1alpha2.FailoverServiceList) FailoverServiceSet {
    list := make([]*networking_mesh_gloo_solo_io_v1alpha2.FailoverService, 0, len(failoverServiceList.Items))
    for idx := range failoverServiceList.Items {
        list = append(list, &failoverServiceList.Items[idx])
    }
    return &failoverServiceSet{set: makeGenericFailoverServiceSet(list)}
}

func (s *failoverServiceSet) Keys() sets.String {
	if s == nil {
		return sets.String{}
    }
    return s.set.Keys()
}

func (s *failoverServiceSet) List(filterResource ... func(*networking_mesh_gloo_solo_io_v1alpha2.FailoverService) bool) []*networking_mesh_gloo_solo_io_v1alpha2.FailoverService {
    if s == nil {
        return nil
    }
    var genericFilters []func(ezkube.ResourceId) bool
    for _, filter := range filterResource {
        genericFilters = append(genericFilters, func(obj ezkube.ResourceId) bool {
            return filter(obj.(*networking_mesh_gloo_solo_io_v1alpha2.FailoverService))
        })
    }

    var failoverServiceList []*networking_mesh_gloo_solo_io_v1alpha2.FailoverService
    for _, obj := range s.set.List(genericFilters...) {
        failoverServiceList = append(failoverServiceList, obj.(*networking_mesh_gloo_solo_io_v1alpha2.FailoverService))
    }
    return failoverServiceList
}

func (s *failoverServiceSet) Map() map[string]*networking_mesh_gloo_solo_io_v1alpha2.FailoverService {
    if s == nil {
        return nil
    }

    newMap := map[string]*networking_mesh_gloo_solo_io_v1alpha2.FailoverService{}
    for k, v := range s.set.Map() {
        newMap[k] = v.(*networking_mesh_gloo_solo_io_v1alpha2.FailoverService)
    }
    return newMap
}

func (s *failoverServiceSet) Insert(
        failoverServiceList ...*networking_mesh_gloo_solo_io_v1alpha2.FailoverService,
) {
    if s == nil {
        panic("cannot insert into nil set")
    }

    for _, obj := range failoverServiceList {
        s.set.Insert(obj)
    }
}

func (s *failoverServiceSet) Has(failoverService ezkube.ResourceId) bool {
    if s == nil {
        return false
    }
    return s.set.Has(failoverService)
}

func (s *failoverServiceSet) Equal(
        failoverServiceSet FailoverServiceSet,
) bool {
    if s == nil {
        return failoverServiceSet == nil
    }
    return s.set.Equal(makeGenericFailoverServiceSet(failoverServiceSet.List()))
}

func (s *failoverServiceSet) Delete(FailoverService ezkube.ResourceId) {
    if s == nil {
        return
    }
    s.set.Delete(FailoverService)
}

func (s *failoverServiceSet) Union(set FailoverServiceSet) FailoverServiceSet {
    if s == nil {
        return set
    }
    return NewFailoverServiceSet(append(s.List(), set.List()...)...)
}

func (s *failoverServiceSet) Difference(set FailoverServiceSet) FailoverServiceSet {
    if s == nil {
        return set
    }
    newSet := s.set.Difference(makeGenericFailoverServiceSet(set.List()))
    return &failoverServiceSet{set: newSet}
}

func (s *failoverServiceSet) Intersection(set FailoverServiceSet) FailoverServiceSet {
    if s == nil {
        return nil
    }
    newSet := s.set.Intersection(makeGenericFailoverServiceSet(set.List()))
    var failoverServiceList []*networking_mesh_gloo_solo_io_v1alpha2.FailoverService
    for _, obj := range newSet.List() {
        failoverServiceList = append(failoverServiceList, obj.(*networking_mesh_gloo_solo_io_v1alpha2.FailoverService))
    }
    return NewFailoverServiceSet(failoverServiceList...)
}


func (s *failoverServiceSet) Find(id ezkube.ResourceId) (*networking_mesh_gloo_solo_io_v1alpha2.FailoverService, error) {
    if s == nil {
        return nil, eris.Errorf("empty set, cannot find FailoverService %v", sksets.Key(id))
    }
	obj, err := s.set.Find(&networking_mesh_gloo_solo_io_v1alpha2.FailoverService{}, id)
	if err != nil {
		return nil, err
    }

    return obj.(*networking_mesh_gloo_solo_io_v1alpha2.FailoverService), nil
}

func (s *failoverServiceSet) Length() int {
    if s == nil {
        return 0
    }
    return s.set.Length()
}
