//go:generate mockgen -source ./snapshot.go -destination mocks/snapshot.go

// Definitions for Output Snapshots
package smi

import (
	"context"
	"encoding/json"
	"sort"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/pkg/multicluster"

	"github.com/rotisserie/eris"
	"github.com/solo-io/skv2/contrib/pkg/output"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	"sigs.k8s.io/controller-runtime/pkg/client"

	split_smi_spec_io_v1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha2"
	split_smi_spec_io_v1alpha2_sets "github.com/solo-io/external-apis/pkg/api/smi/split.smi-spec.io/v1alpha2/sets"

	access_smi_spec_io_v1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/access/v1alpha2"
	access_smi_spec_io_v1alpha2_sets "github.com/solo-io/external-apis/pkg/api/smi/access.smi-spec.io/v1alpha2/sets"

	specs_smi_spec_io_v1alpha3 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/specs/v1alpha3"
	specs_smi_spec_io_v1alpha3_sets "github.com/solo-io/external-apis/pkg/api/smi/specs.smi-spec.io/v1alpha3/sets"
)

// this error can occur if constructing a Partitioned Snapshot from a resource
// that is missing the partition label
var MissingRequiredLabelError = func(labelKey, resourceKind string, obj ezkube.ResourceId) error {
	return eris.Errorf("expected label %v not on labels of %v %v", labelKey, resourceKind, sets.Key(obj))
}

// the snapshot of output resources produced by a translation
type Snapshot interface {

	// return the set of SplitSmiSpecIov1Alpha2TrafficSplits with a given set of labels
	SplitSmiSpecIov1Alpha2TrafficSplits() []LabeledSplitSmiSpecIov1Alpha2TrafficSplitSet
	// return the set of AccessSmiSpecIov1Alpha2TrafficTargets with a given set of labels
	AccessSmiSpecIov1Alpha2TrafficTargets() []LabeledAccessSmiSpecIov1Alpha2TrafficTargetSet
	// return the set of SpecsSmiSpecIov1Alpha3HTTPRouteGroups with a given set of labels
	SpecsSmiSpecIov1Alpha3HTTPRouteGroups() []LabeledSpecsSmiSpecIov1Alpha3HTTPRouteGroupSet

	// apply the snapshot to the local cluster, garbage collecting stale resources
	ApplyLocalCluster(ctx context.Context, clusterClient client.Client, errHandler output.ErrorHandler)

	// apply resources from the snapshot across multiple clusters, garbage collecting stale resources
	ApplyMultiCluster(ctx context.Context, multiClusterClient multicluster.Client, errHandler output.ErrorHandler)

	// serialize the entire snapshot as JSON
	MarshalJSON() ([]byte, error)
}

type snapshot struct {
	name string

	splitSmiSpecIov1Alpha2TrafficSplits   []LabeledSplitSmiSpecIov1Alpha2TrafficSplitsSet
	accessSmiSpecIov1Alpha2TrafficTargets []LabeledAccessSmiSpecIov1Alpha2TrafficTargetsSet
	specsSmiSpecIov1Alpha3HTTPRouteGroups []LabeledSpecsSmiSpecIov1Alpha3HTTPRouteGroupsSet
	clusters                              []string
}

func NewSnapshot(
	name string,

	splitSmiSpecIov1Alpha2TrafficSplits []LabeledSplitSmiSpecIov1Alpha2TrafficSplitsSet,
	accessSmiSpecIov1Alpha2TrafficTargets []LabeledAccessSmiSpecIov1Alpha2TrafficTargetsSet,
	specsSmiSpecIov1Alpha3HTTPRouteGroups []LabeledSpecsSmiSpecIov1Alpha3HTTPRouteGroupsSet,
	clusters ...string, // the set of clusters to apply the snapshot to. only required for multicluster snapshots.
) Snapshot {
	return &snapshot{
		name: name,

		splitSmiSpecIov1Alpha2TrafficSplits:   splitSmiSpecIov1Alpha2TrafficSplits,
		accessSmiSpecIov1Alpha2TrafficTargets: accessSmiSpecIov1Alpha2TrafficTargets,
		specsSmiSpecIov1Alpha3HTTPRouteGroups: specsSmiSpecIov1Alpha3HTTPRouteGroups,
		clusters:                              clusters,
	}
}

// automatically partitions the input resources
// by the presence of the provided label.
func NewLabelPartitionedSnapshot(
	name,
	labelKey string, // the key by which to partition the resources

	splitSmiSpecIov1Alpha2TrafficSplits split_smi_spec_io_v1alpha2_sets.TrafficSplitSet,

	accessSmiSpecIov1Alpha2TrafficTargets access_smi_spec_io_v1alpha2_sets.TrafficTargetSet,

	specsSmiSpecIov1Alpha3HTTPRouteGroups specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet,
	clusters ...string, // the set of clusters to apply the snapshot to. only required for multicluster snapshots.
) (Snapshot, error) {

	partitionedSplitSmiSpecIov1Alpha2TrafficSplits, err := partitionSplitSmiSpecIov1Alpha2TrafficSplitsByLabel(labelKey, splitSmiSpecIov1Alpha2TrafficSplits)
	if err != nil {
		return nil, err
	}
	partitionedAccessSmiSpecIov1Alpha2TrafficTargets, err := partitionAccessSmiSpecIov1Alpha2TrafficTargetsByLabel(labelKey, accessSmiSpecIov1Alpha2TrafficTargets)
	if err != nil {
		return nil, err
	}
	partitionedSpecsSmiSpecIov1Alpha3HTTPRouteGroups, err := partitionSpecsSmiSpecIov1Alpha3HTTPRouteGroupsByLabel(labelKey, specsSmiSpecIov1Alpha3HTTPRouteGroups)
	if err != nil {
		return nil, err
	}

	return NewSnapshot(
		name,

		partitionedSplitSmiSpecIov1Alpha2TrafficSplits,
		partitionedAccessSmiSpecIov1Alpha2TrafficTargets,
		partitionedSpecsSmiSpecIov1Alpha3HTTPRouteGroups,
		clusters...,
	), nil
}

// simplified constructor for a snapshot
// with a single label partition (i.e. all resources share a single set of labels).
func NewSinglePartitionedSnapshot(
	name string,
	snapshotLabels map[string]string, // a single set of labels shared by all resources

	splitSmiSpecIov1Alpha2TrafficSplits split_smi_spec_io_v1alpha2_sets.TrafficSplitSet,

	accessSmiSpecIov1Alpha2TrafficTargets access_smi_spec_io_v1alpha2_sets.TrafficTargetSet,

	specsSmiSpecIov1Alpha3HTTPRouteGroups specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet,
	clusters ...string, // the set of clusters to apply the snapshot to. only required for multicluster snapshots.
) (Snapshot, error) {

	labeledSplitSmiSpecIov1Alpha2TrafficSplit, err := NewLabeledSplitSmiSpecIov1Alpha2TrafficSplitSet(splitSmiSpecIov1Alpha2TrafficSplits, snapshotLabels)
	if err != nil {
		return nil, err
	}
	labeledAccessSmiSpecIov1Alpha2TrafficTarget, err := NewLabeledAccessSmiSpecIov1Alpha2TrafficTargetSet(accessSmiSpecIov1Alpha2TrafficTargets, snapshotLabels)
	if err != nil {
		return nil, err
	}
	labeledSpecsSmiSpecIov1Alpha3HTTPRouteGroup, err := NewLabeledSpecsSmiSpecIov1Alpha3HTTPRouteGroupSet(specsSmiSpecIov1Alpha3HTTPRouteGroups, snapshotLabels)
	if err != nil {
		return nil, err
	}

	return NewSnapshot(
		name,

		[]LabeledSplitSmiSpecIov1Alpha2TrafficSplitSet{labeledSplitSmiSpecIov1Alpha2TrafficSplit},
		[]LabeledAccessSmiSpecIov1Alpha2TrafficTargetSet{labeledAccessSmiSpecIov1Alpha2TrafficTarget},
		[]LabeledSpecsSmiSpecIov1Alpha3HTTPRouteGroupSet{labeledSpecsSmiSpecIov1Alpha3HTTPRouteGroup},
		clusters...,
	), nil
}

// apply the desired resources to the cluster state; remove stale resources where necessary
func (s *snapshot) ApplyLocalCluster(ctx context.Context, cli client.Client, errHandler output.ErrorHandler) {
	var genericLists []output.ResourceList

	for _, outputSet := range s.splitSmiSpecIov1Alpha2TrafficSplits {
		genericLists = append(genericLists, outputSet.Generic())
	}
	for _, outputSet := range s.accessSmiSpecIov1Alpha2TrafficTargets {
		genericLists = append(genericLists, outputSet.Generic())
	}
	for _, outputSet := range s.specsSmiSpecIov1Alpha3HTTPRouteGroups {
		genericLists = append(genericLists, outputSet.Generic())
	}

	output.Snapshot{
		Name:        s.name,
		ListsToSync: genericLists,
	}.SyncLocalCluster(ctx, cli, errHandler)
}

// apply the desired resources to multiple cluster states; remove stale resources where necessary
func (s *snapshot) ApplyMultiCluster(ctx context.Context, multiClusterClient multicluster.Client, errHandler output.ErrorHandler) {
	var genericLists []output.ResourceList

	for _, outputSet := range s.splitSmiSpecIov1Alpha2TrafficSplits {
		genericLists = append(genericLists, outputSet.Generic())
	}
	for _, outputSet := range s.accessSmiSpecIov1Alpha2TrafficTargets {
		genericLists = append(genericLists, outputSet.Generic())
	}
	for _, outputSet := range s.specsSmiSpecIov1Alpha3HTTPRouteGroups {
		genericLists = append(genericLists, outputSet.Generic())
	}

	output.Snapshot{
		Name:        s.name,
		Clusters:    s.clusters,
		ListsToSync: genericLists,
	}.SyncMultiCluster(ctx, multiClusterClient, errHandler)
}

func partitionSplitSmiSpecIov1Alpha2TrafficSplitsByLabel(labelKey string, set split_smi_spec_io_v1alpha2_sets.TrafficSplitSet) ([]LabeledSplitSmiSpecIov1Alpha2TrafficSplitSet, error) {
	setsByLabel := map[string]split_smi_spec_io_v1alpha2_sets.TrafficSplitSet{}

	for _, obj := range set.List() {
		if obj.Labels == nil {
			return nil, MissingRequiredLabelError(labelKey, "SplitSmiSpecIov1Alpha2TrafficSplit", obj)
		}
		labelValue := obj.Labels[labelKey]
		if labelValue == "" {
			return nil, MissingRequiredLabelError(labelKey, "SplitSmiSpecIov1Alpha2TrafficSplit", obj)
		}

		setForValue, ok := setsByLabel[labelValue]
		if !ok {
			setForValue = split_smi_spec_io_v1alpha2_sets.NewTrafficSplitSet()
			setsByLabel[labelValue] = setForValue
		}
		setForValue.Insert(obj)
	}

	// partition by label key
	var partitionedSplitSmiSpecIov1Alpha2TrafficSplits []LabeledSplitSmiSpecIov1Alpha2TrafficSplitSet

	for labelValue, setForValue := range setsByLabel {
		labels := map[string]string{labelKey: labelValue}

		partitionedSet, err := NewLabeledSplitSmiSpecIov1Alpha2TrafficSplitSet(setForValue, labels)
		if err != nil {
			return nil, err
		}

		partitionedSplitSmiSpecIov1Alpha2TrafficSplits = append(partitionedSplitSmiSpecIov1Alpha2TrafficSplits, partitionedSet)
	}

	// sort for idempotency
	sort.SliceStable(partitionedSplitSmiSpecIov1Alpha2TrafficSplits, func(i, j int) bool {
		leftLabelValue := partitionedSplitSmiSpecIov1Alpha2TrafficSplits[i].Labels()[labelKey]
		rightLabelValue := partitionedSplitSmiSpecIov1Alpha2TrafficSplits[j].Labels()[labelKey]
		return leftLabelValue < rightLabelValue
	})

	return partitionedSplitSmiSpecIov1Alpha2TrafficSplits, nil
}

func partitionAccessSmiSpecIov1Alpha2TrafficTargetsByLabel(labelKey string, set access_smi_spec_io_v1alpha2_sets.TrafficTargetSet) ([]LabeledAccessSmiSpecIov1Alpha2TrafficTargetSet, error) {
	setsByLabel := map[string]access_smi_spec_io_v1alpha2_sets.TrafficTargetSet{}

	for _, obj := range set.List() {
		if obj.Labels == nil {
			return nil, MissingRequiredLabelError(labelKey, "AccessSmiSpecIov1Alpha2TrafficTarget", obj)
		}
		labelValue := obj.Labels[labelKey]
		if labelValue == "" {
			return nil, MissingRequiredLabelError(labelKey, "AccessSmiSpecIov1Alpha2TrafficTarget", obj)
		}

		setForValue, ok := setsByLabel[labelValue]
		if !ok {
			setForValue = access_smi_spec_io_v1alpha2_sets.NewTrafficTargetSet()
			setsByLabel[labelValue] = setForValue
		}
		setForValue.Insert(obj)
	}

	// partition by label key
	var partitionedAccessSmiSpecIov1Alpha2TrafficTargets []LabeledAccessSmiSpecIov1Alpha2TrafficTargetSet

	for labelValue, setForValue := range setsByLabel {
		labels := map[string]string{labelKey: labelValue}

		partitionedSet, err := NewLabeledAccessSmiSpecIov1Alpha2TrafficTargetSet(setForValue, labels)
		if err != nil {
			return nil, err
		}

		partitionedAccessSmiSpecIov1Alpha2TrafficTargets = append(partitionedAccessSmiSpecIov1Alpha2TrafficTargets, partitionedSet)
	}

	// sort for idempotency
	sort.SliceStable(partitionedAccessSmiSpecIov1Alpha2TrafficTargets, func(i, j int) bool {
		leftLabelValue := partitionedAccessSmiSpecIov1Alpha2TrafficTargets[i].Labels()[labelKey]
		rightLabelValue := partitionedAccessSmiSpecIov1Alpha2TrafficTargets[j].Labels()[labelKey]
		return leftLabelValue < rightLabelValue
	})

	return partitionedAccessSmiSpecIov1Alpha2TrafficTargets, nil
}

func partitionSpecsSmiSpecIov1Alpha3HTTPRouteGroupsByLabel(labelKey string, set specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet) ([]LabeledSpecsSmiSpecIov1Alpha3HTTPRouteGroupSet, error) {
	setsByLabel := map[string]specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet{}

	for _, obj := range set.List() {
		if obj.Labels == nil {
			return nil, MissingRequiredLabelError(labelKey, "SpecsSmiSpecIov1Alpha3HTTPRouteGroup", obj)
		}
		labelValue := obj.Labels[labelKey]
		if labelValue == "" {
			return nil, MissingRequiredLabelError(labelKey, "SpecsSmiSpecIov1Alpha3HTTPRouteGroup", obj)
		}

		setForValue, ok := setsByLabel[labelValue]
		if !ok {
			setForValue = specs_smi_spec_io_v1alpha3_sets.NewHTTPRouteGroupSet()
			setsByLabel[labelValue] = setForValue
		}
		setForValue.Insert(obj)
	}

	// partition by label key
	var partitionedSpecsSmiSpecIov1Alpha3HTTPRouteGroups []LabeledSpecsSmiSpecIov1Alpha3HTTPRouteGroupSet

	for labelValue, setForValue := range setsByLabel {
		labels := map[string]string{labelKey: labelValue}

		partitionedSet, err := NewLabeledSpecsSmiSpecIov1Alpha3HTTPRouteGroupSet(setForValue, labels)
		if err != nil {
			return nil, err
		}

		partitionedSpecsSmiSpecIov1Alpha3HTTPRouteGroups = append(partitionedSpecsSmiSpecIov1Alpha3HTTPRouteGroups, partitionedSet)
	}

	// sort for idempotency
	sort.SliceStable(partitionedSpecsSmiSpecIov1Alpha3HTTPRouteGroups, func(i, j int) bool {
		leftLabelValue := partitionedSpecsSmiSpecIov1Alpha3HTTPRouteGroups[i].Labels()[labelKey]
		rightLabelValue := partitionedSpecsSmiSpecIov1Alpha3HTTPRouteGroups[j].Labels()[labelKey]
		return leftLabelValue < rightLabelValue
	})

	return partitionedSpecsSmiSpecIov1Alpha3HTTPRouteGroups, nil
}

func (s snapshot) SplitSmiSpecIov1Alpha2TrafficSplits() []LabeledSplitSmiSpecIov1Alpha2TrafficSplitSet {
	return s.splitSmiSpecIov1Alpha2TrafficSplits
}

func (s snapshot) AccessSmiSpecIov1Alpha2TrafficTargets() []LabeledAccessSmiSpecIov1Alpha2TrafficTargetSet {
	return s.accessSmiSpecIov1Alpha2TrafficTargets
}

func (s snapshot) SpecsSmiSpecIov1Alpha3HTTPRouteGroups() []LabeledSpecsSmiSpecIov1Alpha3HTTPRouteGroupSet {
	return s.specsSmiSpecIov1Alpha3HTTPRouteGroups
}

func (s snapshot) MarshalJSON() ([]byte, error) {
	snapshotMap := map[string]interface{}{"name": s.name}

	splitSmiSpecIov1Alpha2TrafficSplitSet := split_smi_spec_io_v1alpha2_sets.NewTrafficSplitSet()
	for _, set := range s.splitSmiSpecIov1Alpha2TrafficSplits {
		splitSmiSpecIov1Alpha2TrafficSplitSet = splitSmiSpecIov1Alpha2TrafficSplitSet.Union(set.Set())
	}
	snapshotMap["splitSmiSpecIov1Alpha2TrafficSplits"] = splitSmiSpecIov1Alpha2TrafficSplitSet.List()

	accessSmiSpecIov1Alpha2TrafficTargetSet := access_smi_spec_io_v1alpha2_sets.NewTrafficTargetSet()
	for _, set := range s.accessSmiSpecIov1Alpha2TrafficTargets {
		accessSmiSpecIov1Alpha2TrafficTargetSet = accessSmiSpecIov1Alpha2TrafficTargetSet.Union(set.Set())
	}
	snapshotMap["accessSmiSpecIov1Alpha2TrafficTargets"] = accessSmiSpecIov1Alpha2TrafficTargetSet.List()

	specsSmiSpecIov1Alpha3HTTPRouteGroupSet := specs_smi_spec_io_v1alpha3_sets.NewHTTPRouteGroupSet()
	for _, set := range s.specsSmiSpecIov1Alpha3HTTPRouteGroups {
		specsSmiSpecIov1Alpha3HTTPRouteGroupSet = specsSmiSpecIov1Alpha3HTTPRouteGroupSet.Union(set.Set())
	}
	snapshotMap["specsSmiSpecIov1Alpha3HTTPRouteGroups"] = specsSmiSpecIov1Alpha3HTTPRouteGroupSet.List()

	snapshotMap["clusters"] = s.clusters

	return json.Marshal(snapshotMap)
}

// LabeledSplitSmiSpecIov1Alpha2TrafficSplitSet represents a set of splitSmiSpecIov1Alpha2TrafficSplits
// which share a common set of labels.
// These labels are used to find diffs between SplitSmiSpecIov1Alpha2TrafficSplitSets.
type LabeledSplitSmiSpecIov1Alpha2TrafficSplitSet interface {
	// returns the set of Labels shared by this SplitSmiSpecIov1Alpha2TrafficSplitSet
	Labels() map[string]string

	// returns the set of TrafficSplites with the given labels
	Set() split_smi_spec_io_v1alpha2_sets.TrafficSplitSet

	// converts the set to a generic format which can be applied by the Snapshot.Apply functions
	Generic() output.ResourceList
}

type labeledSplitSmiSpecIov1Alpha2TrafficSplitSet struct {
	set    split_smi_spec_io_v1alpha2_sets.TrafficSplitSet
	labels map[string]string
}

func NewLabeledSplitSmiSpecIov1Alpha2TrafficSplitSet(set split_smi_spec_io_v1alpha2_sets.TrafficSplitSet, labels map[string]string) (LabeledSplitSmiSpecIov1Alpha2TrafficSplitSet, error) {
	// validate that each TrafficSplit contains the labels, else this is not a valid LabeledSplitSmiSpecIov1Alpha2TrafficSplitSet
	for _, item := range set.List() {
		for k, v := range labels {
			// k=v must be present in the item
			if item.Labels[k] != v {
				return nil, eris.Errorf("internal error: %v=%v missing on SplitSmiSpecIov1Alpha2TrafficSplit %v", k, v, item.Name)
			}
		}
	}

	return &labeledSplitSmiSpecIov1Alpha2TrafficSplitSet{set: set, labels: labels}, nil
}

func (l *labeledSplitSmiSpecIov1Alpha2TrafficSplitSet) Labels() map[string]string {
	return l.labels
}

func (l *labeledSplitSmiSpecIov1Alpha2TrafficSplitSet) Set() split_smi_spec_io_v1alpha2_sets.TrafficSplitSet {
	return l.set
}

func (l labeledSplitSmiSpecIov1Alpha2TrafficSplitSet) Generic() output.ResourceList {
	var desiredResources []ezkube.Object
	for _, desired := range l.set.List() {
		desiredResources = append(desiredResources, desired)
	}

	// enable list func for garbage collection
	listFunc := func(ctx context.Context, cli client.Client) ([]ezkube.Object, error) {
		var list split_smi_spec_io_v1alpha2.TrafficSplitList
		if err := cli.List(ctx, &list, client.MatchingLabels(l.labels)); err != nil {
			return nil, err
		}
		var items []ezkube.Object
		for _, item := range list.Items {
			item := item // pike
			items = append(items, &item)
		}
		return items, nil
	}

	return output.ResourceList{
		Resources:    desiredResources,
		ListFunc:     listFunc,
		ResourceKind: "TrafficSplit",
	}
}

// LabeledAccessSmiSpecIov1Alpha2TrafficTargetSet represents a set of accessSmiSpecIov1Alpha2TrafficTargets
// which share a common set of labels.
// These labels are used to find diffs between AccessSmiSpecIov1Alpha2TrafficTargetSets.
type LabeledAccessSmiSpecIov1Alpha2TrafficTargetSet interface {
	// returns the set of Labels shared by this AccessSmiSpecIov1Alpha2TrafficTargetSet
	Labels() map[string]string

	// returns the set of TrafficTargetes with the given labels
	Set() access_smi_spec_io_v1alpha2_sets.TrafficTargetSet

	// converts the set to a generic format which can be applied by the Snapshot.Apply functions
	Generic() output.ResourceList
}

type labeledAccessSmiSpecIov1Alpha2TrafficTargetSet struct {
	set    access_smi_spec_io_v1alpha2_sets.TrafficTargetSet
	labels map[string]string
}

func NewLabeledAccessSmiSpecIov1Alpha2TrafficTargetSet(set access_smi_spec_io_v1alpha2_sets.TrafficTargetSet, labels map[string]string) (LabeledAccessSmiSpecIov1Alpha2TrafficTargetSet, error) {
	// validate that each TrafficTarget contains the labels, else this is not a valid LabeledAccessSmiSpecIov1Alpha2TrafficTargetSet
	for _, item := range set.List() {
		for k, v := range labels {
			// k=v must be present in the item
			if item.Labels[k] != v {
				return nil, eris.Errorf("internal error: %v=%v missing on AccessSmiSpecIov1Alpha2TrafficTarget %v", k, v, item.Name)
			}
		}
	}

	return &labeledAccessSmiSpecIov1Alpha2TrafficTargetSet{set: set, labels: labels}, nil
}

func (l *labeledAccessSmiSpecIov1Alpha2TrafficTargetSet) Labels() map[string]string {
	return l.labels
}

func (l *labeledAccessSmiSpecIov1Alpha2TrafficTargetSet) Set() access_smi_spec_io_v1alpha2_sets.TrafficTargetSet {
	return l.set
}

func (l labeledAccessSmiSpecIov1Alpha2TrafficTargetSet) Generic() output.ResourceList {
	var desiredResources []ezkube.Object
	for _, desired := range l.set.List() {
		desiredResources = append(desiredResources, desired)
	}

	// enable list func for garbage collection
	listFunc := func(ctx context.Context, cli client.Client) ([]ezkube.Object, error) {
		var list access_smi_spec_io_v1alpha2.TrafficTargetList
		if err := cli.List(ctx, &list, client.MatchingLabels(l.labels)); err != nil {
			return nil, err
		}
		var items []ezkube.Object
		for _, item := range list.Items {
			item := item // pike
			items = append(items, &item)
		}
		return items, nil
	}

	return output.ResourceList{
		Resources:    desiredResources,
		ListFunc:     listFunc,
		ResourceKind: "TrafficTarget",
	}
}

// LabeledSpecsSmiSpecIov1Alpha3HTTPRouteGroupSet represents a set of specsSmiSpecIov1Alpha3HTTPRouteGroups
// which share a common set of labels.
// These labels are used to find diffs between SpecsSmiSpecIov1Alpha3HTTPRouteGroupSets.
type LabeledSpecsSmiSpecIov1Alpha3HTTPRouteGroupSet interface {
	// returns the set of Labels shared by this SpecsSmiSpecIov1Alpha3HTTPRouteGroupSet
	Labels() map[string]string

	// returns the set of HTTPRouteGroupes with the given labels
	Set() specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet

	// converts the set to a generic format which can be applied by the Snapshot.Apply functions
	Generic() output.ResourceList
}

type labeledSpecsSmiSpecIov1Alpha3HTTPRouteGroupSet struct {
	set    specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet
	labels map[string]string
}

func NewLabeledSpecsSmiSpecIov1Alpha3HTTPRouteGroupSet(set specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet, labels map[string]string) (LabeledSpecsSmiSpecIov1Alpha3HTTPRouteGroupSet, error) {
	// validate that each HTTPRouteGroup contains the labels, else this is not a valid LabeledSpecsSmiSpecIov1Alpha3HTTPRouteGroupSet
	for _, item := range set.List() {
		for k, v := range labels {
			// k=v must be present in the item
			if item.Labels[k] != v {
				return nil, eris.Errorf("internal error: %v=%v missing on SpecsSmiSpecIov1Alpha3HTTPRouteGroup %v", k, v, item.Name)
			}
		}
	}

	return &labeledSpecsSmiSpecIov1Alpha3HTTPRouteGroupSet{set: set, labels: labels}, nil
}

func (l *labeledSpecsSmiSpecIov1Alpha3HTTPRouteGroupSet) Labels() map[string]string {
	return l.labels
}

func (l *labeledSpecsSmiSpecIov1Alpha3HTTPRouteGroupSet) Set() specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet {
	return l.set
}

func (l labeledSpecsSmiSpecIov1Alpha3HTTPRouteGroupSet) Generic() output.ResourceList {
	var desiredResources []ezkube.Object
	for _, desired := range l.set.List() {
		desiredResources = append(desiredResources, desired)
	}

	// enable list func for garbage collection
	listFunc := func(ctx context.Context, cli client.Client) ([]ezkube.Object, error) {
		var list specs_smi_spec_io_v1alpha3.HTTPRouteGroupList
		if err := cli.List(ctx, &list, client.MatchingLabels(l.labels)); err != nil {
			return nil, err
		}
		var items []ezkube.Object
		for _, item := range list.Items {
			item := item // pike
			items = append(items, &item)
		}
		return items, nil
	}

	return output.ResourceList{
		Resources:    desiredResources,
		ListFunc:     listFunc,
		ResourceKind: "HTTPRouteGroup",
	}
}

type builder struct {
	ctx      context.Context
	name     string
	clusters []string

	splitSmiSpecIov1Alpha2TrafficSplits split_smi_spec_io_v1alpha2_sets.TrafficSplitSet

	accessSmiSpecIov1Alpha2TrafficTargets access_smi_spec_io_v1alpha2_sets.TrafficTargetSet

	specsSmiSpecIov1Alpha3HTTPRouteGroups specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet
}

func NewBuilder(ctx context.Context, name string) *builder {
	return &builder{
		ctx:  ctx,
		name: name,

		splitSmiSpecIov1Alpha2TrafficSplits: split_smi_spec_io_v1alpha2_sets.NewTrafficSplitSet(),

		accessSmiSpecIov1Alpha2TrafficTargets: access_smi_spec_io_v1alpha2_sets.NewTrafficTargetSet(),

		specsSmiSpecIov1Alpha3HTTPRouteGroups: specs_smi_spec_io_v1alpha3_sets.NewHTTPRouteGroupSet(),
	}
}

// the output Builder uses a builder pattern to allow
// iteratively collecting outputs before producing a final snapshot
type Builder interface {

	// add SplitSmiSpecIov1Alpha2TrafficSplits to the collected outputs
	AddSplitSmiSpecIov1Alpha2TrafficSplits(splitSmiSpecIov1Alpha2TrafficSplits ...*split_smi_spec_io_v1alpha2.TrafficSplit)

	// get the collected SplitSmiSpecIov1Alpha2TrafficSplits
	GetSplitSmiSpecIov1Alpha2TrafficSplits() split_smi_spec_io_v1alpha2_sets.TrafficSplitSet

	// add AccessSmiSpecIov1Alpha2TrafficTargets to the collected outputs
	AddAccessSmiSpecIov1Alpha2TrafficTargets(accessSmiSpecIov1Alpha2TrafficTargets ...*access_smi_spec_io_v1alpha2.TrafficTarget)

	// get the collected AccessSmiSpecIov1Alpha2TrafficTargets
	GetAccessSmiSpecIov1Alpha2TrafficTargets() access_smi_spec_io_v1alpha2_sets.TrafficTargetSet

	// add SpecsSmiSpecIov1Alpha3HTTPRouteGroups to the collected outputs
	AddSpecsSmiSpecIov1Alpha3HTTPRouteGroups(specsSmiSpecIov1Alpha3HTTPRouteGroups ...*specs_smi_spec_io_v1alpha3.HTTPRouteGroup)

	// get the collected SpecsSmiSpecIov1Alpha3HTTPRouteGroups
	GetSpecsSmiSpecIov1Alpha3HTTPRouteGroups() specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet

	// build the collected outputs into a label-partitioned snapshot
	BuildLabelPartitionedSnapshot(labelKey string) (Snapshot, error)

	// build the collected outputs into a snapshot with a single partition
	BuildSinglePartitionedSnapshot(snapshotLabels map[string]string) (Snapshot, error)

	// add a cluster to the collected clusters.
	// this can be used to collect clusters for use with MultiCluster snapshots.
	AddCluster(cluster string)

	// returns the set of clusters currently stored in this builder
	Clusters() []string

	// merge all the resources from another Builder into this one
	Merge(other Builder)

	// create a clone of this builder (deepcopying all resources)
	Clone() Builder

	// return the difference between the snapshot in this builder's and another
	Delta(newSnap Builder) output.SnapshotDelta
}

func (b *builder) AddSplitSmiSpecIov1Alpha2TrafficSplits(splitSmiSpecIov1Alpha2TrafficSplits ...*split_smi_spec_io_v1alpha2.TrafficSplit) {
	for _, obj := range splitSmiSpecIov1Alpha2TrafficSplits {
		if obj == nil {
			continue
		}
		contextutils.LoggerFrom(b.ctx).Debugf("added output SplitSmiSpecIov1Alpha2TrafficSplit %v", sets.Key(obj))
		b.splitSmiSpecIov1Alpha2TrafficSplits.Insert(obj)
	}
}
func (b *builder) AddAccessSmiSpecIov1Alpha2TrafficTargets(accessSmiSpecIov1Alpha2TrafficTargets ...*access_smi_spec_io_v1alpha2.TrafficTarget) {
	for _, obj := range accessSmiSpecIov1Alpha2TrafficTargets {
		if obj == nil {
			continue
		}
		contextutils.LoggerFrom(b.ctx).Debugf("added output AccessSmiSpecIov1Alpha2TrafficTarget %v", sets.Key(obj))
		b.accessSmiSpecIov1Alpha2TrafficTargets.Insert(obj)
	}
}
func (b *builder) AddSpecsSmiSpecIov1Alpha3HTTPRouteGroups(specsSmiSpecIov1Alpha3HTTPRouteGroups ...*specs_smi_spec_io_v1alpha3.HTTPRouteGroup) {
	for _, obj := range specsSmiSpecIov1Alpha3HTTPRouteGroups {
		if obj == nil {
			continue
		}
		contextutils.LoggerFrom(b.ctx).Debugf("added output SpecsSmiSpecIov1Alpha3HTTPRouteGroup %v", sets.Key(obj))
		b.specsSmiSpecIov1Alpha3HTTPRouteGroups.Insert(obj)
	}
}

func (b *builder) GetSplitSmiSpecIov1Alpha2TrafficSplits() split_smi_spec_io_v1alpha2_sets.TrafficSplitSet {
	return b.splitSmiSpecIov1Alpha2TrafficSplits
}

func (b *builder) GetAccessSmiSpecIov1Alpha2TrafficTargets() access_smi_spec_io_v1alpha2_sets.TrafficTargetSet {
	return b.accessSmiSpecIov1Alpha2TrafficTargets
}

func (b *builder) GetSpecsSmiSpecIov1Alpha3HTTPRouteGroups() specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet {
	return b.specsSmiSpecIov1Alpha3HTTPRouteGroups
}

func (b *builder) BuildLabelPartitionedSnapshot(labelKey string) (Snapshot, error) {
	return NewLabelPartitionedSnapshot(
		b.name,
		labelKey,

		b.splitSmiSpecIov1Alpha2TrafficSplits,

		b.accessSmiSpecIov1Alpha2TrafficTargets,

		b.specsSmiSpecIov1Alpha3HTTPRouteGroups,
		b.clusters...,
	)
}

func (b *builder) BuildSinglePartitionedSnapshot(snapshotLabels map[string]string) (Snapshot, error) {
	return NewSinglePartitionedSnapshot(
		b.name,
		snapshotLabels,

		b.splitSmiSpecIov1Alpha2TrafficSplits,

		b.accessSmiSpecIov1Alpha2TrafficTargets,

		b.specsSmiSpecIov1Alpha3HTTPRouteGroups,
		b.clusters...,
	)
}

func (b *builder) AddCluster(cluster string) {
	b.clusters = append(b.clusters, cluster)
}

func (b *builder) Clusters() []string {
	return b.clusters
}

func (b *builder) Merge(other Builder) {
	if other == nil {
		return
	}

	b.AddSplitSmiSpecIov1Alpha2TrafficSplits(other.GetSplitSmiSpecIov1Alpha2TrafficSplits().List()...)

	b.AddAccessSmiSpecIov1Alpha2TrafficTargets(other.GetAccessSmiSpecIov1Alpha2TrafficTargets().List()...)

	b.AddSpecsSmiSpecIov1Alpha3HTTPRouteGroups(other.GetSpecsSmiSpecIov1Alpha3HTTPRouteGroups().List()...)
	for _, cluster := range other.Clusters() {
		b.AddCluster(cluster)
	}
}

func (b *builder) Clone() Builder {
	if b == nil {
		return nil
	}
	clone := NewBuilder(b.ctx, b.name)

	for _, splitSmiSpecIov1Alpha2TrafficSplit := range b.GetSplitSmiSpecIov1Alpha2TrafficSplits().List() {
		clone.AddSplitSmiSpecIov1Alpha2TrafficSplits(splitSmiSpecIov1Alpha2TrafficSplit.DeepCopy())
	}

	for _, accessSmiSpecIov1Alpha2TrafficTarget := range b.GetAccessSmiSpecIov1Alpha2TrafficTargets().List() {
		clone.AddAccessSmiSpecIov1Alpha2TrafficTargets(accessSmiSpecIov1Alpha2TrafficTarget.DeepCopy())
	}

	for _, specsSmiSpecIov1Alpha3HTTPRouteGroup := range b.GetSpecsSmiSpecIov1Alpha3HTTPRouteGroups().List() {
		clone.AddSpecsSmiSpecIov1Alpha3HTTPRouteGroups(specsSmiSpecIov1Alpha3HTTPRouteGroup.DeepCopy())
	}
	for _, cluster := range b.Clusters() {
		clone.AddCluster(cluster)
	}
	return clone
}

func (b *builder) Delta(other Builder) output.SnapshotDelta {
	delta := output.SnapshotDelta{}
	if b == nil {
		return delta
	}

	// calcualte delta between TrafficSplits
	trafficSplitDelta := b.GetTrafficSplits().Delta(other.GetTrafficSplits())
	trafficSplitGvk := schema.GroupVersionKind{
		Group:   "split.smi-spec.io",
		Version: "v1alpha2",
		Kind:    "TrafficSplit",
	}
	delta.AddInserted(trafficSplitGvk, trafficSplitDelta.Inserted)
	delta.AddRemoved(trafficSplitGvk, trafficSplitDelta.Removed)

	// calcualte delta between TrafficTargets
	trafficTargetDelta := b.GetTrafficTargets().Delta(other.GetTrafficTargets())
	trafficTargetGvk := schema.GroupVersionKind{
		Group:   "access.smi-spec.io",
		Version: "v1alpha2",
		Kind:    "TrafficTarget",
	}
	delta.AddInserted(trafficTargetGvk, trafficTargetDelta.Inserted)
	delta.AddRemoved(trafficTargetGvk, trafficTargetDelta.Removed)

	// calcualte delta between HTTPRouteGroups
	hTTPRouteGroupDelta := b.GetHTTPRouteGroups().Delta(other.GetHTTPRouteGroups())
	hTTPRouteGroupGvk := schema.GroupVersionKind{
		Group:   "specs.smi-spec.io",
		Version: "v1alpha3",
		Kind:    "HTTPRouteGroup",
	}
	delta.AddInserted(hTTPRouteGroupGvk, hTTPRouteGroupDelta.Inserted)
	delta.AddRemoved(hTTPRouteGroupGvk, hTTPRouteGroupDelta.Removed)
	return delta
}
