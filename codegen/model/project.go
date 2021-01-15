package model

import (
	"path/filepath"

	"github.com/solo-io/gloo-mesh/codegen/io"
	"github.com/solo-io/skv2/codegen/model"
	"github.com/solo-io/skv2/contrib"
)

// encapsulates the top level templates command for the entire project
type Project struct {
	// snapshots will be built from the set of API Groups contained here.
	// map key should be the name of the go module containing the API Group.
	// for API Groups defined in the local repository, the go module should be left as empty string ""
	SnapshotApiGroups map[string][]model.Group

	// The set of components to be generated for this Project
	TopLevelComponents []TopLevelComponent
}

// Build the top level templates for the project
func (p Project) TopLevelTemplates() []model.CustomTemplates {
	var templates []model.CustomTemplates
	for _, component := range p.TopLevelComponents {
		templates = append(templates, component.MakeCodegenTemplates(p.SnapshotApiGroups)...)
	}
	return templates
}

// generates an input snapshot, input reconciler, and output snapshot for each
// top-level component. top-level components are defined
// by mapping a given set of inputs to outputs.
type TopLevelComponent struct {
	// path where the generated top-level component's code will be placed.
	// input snapshots will live in <GeneratedCodeRoot>/input/snapshot.go
	// input reconcilers will live in <GeneratedCodeRoot>/input/reconciler.go
	// output snapshots will live in <GeneratedCodeRoot>/output/snapshot.go
	GeneratedCodeRoot string

	// the set of input resources for which to generate a snapshot and reconciler
	// local inptus are read from the local cluster where the controller runs
	LocalInputResources io.Snapshot
	// remote inptus are read from managed cluster registered to the controller cluster
	RemoteInputResources io.Snapshot

	// name of the local snapshot, if the component is hybrid. defaults to Local
	LocalSnapshotName string

	// name of the remote snapshot, if the component is hybrid. defaults to Remote
	RemoteSnapshotName string

	// if this component can run in agent mode, generate a homogenous reconciler which combines both the local and remote resources
	AgentMode bool

	// the set of output resources for which to generate a snapshot
	OutputResources []io.OutputSnapshot
}

func (t TopLevelComponent) MakeCodegenTemplates(snapshotApiGroups map[string][]model.Group) []model.CustomTemplates {
	var topLevelTemplates []model.CustomTemplates

	switch {
	case t.LocalInputResources != nil && t.RemoteInputResources != nil:
		if t.LocalSnapshotName == "" {
			t.LocalSnapshotName = "Local"
		}
		if t.RemoteSnapshotName == "" {
			t.RemoteSnapshotName = "Remote"
		}
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputSnapshot,
			t.LocalSnapshotName,
			t.GeneratedCodeRoot+"/input/local_snapshot.go",
			contrib.HomogenousSnapshotResources{ResourcesToSelect: t.LocalInputResources},
			snapshotApiGroups,
		))
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputSnapshot,
			t.RemoteSnapshotName,
			t.GeneratedCodeRoot+"/input/remote_snapshot.go",
			contrib.HomogenousSnapshotResources{ResourcesToSelect: t.RemoteInputResources},
			snapshotApiGroups,
		))
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputSnapshotManualBuilder,
			t.LocalSnapshotName,
			t.GeneratedCodeRoot+"/input/local_snapshot_manual_builder.go",
			contrib.HomogenousSnapshotResources{ResourcesToSelect: t.LocalInputResources},
			snapshotApiGroups,
		))
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputSnapshotManualBuilder,
			t.RemoteSnapshotName,
			t.GeneratedCodeRoot+"/input/remote_snapshot_manual_builder.go",
			contrib.HomogenousSnapshotResources{ResourcesToSelect: t.RemoteInputResources},
			snapshotApiGroups,
		))

		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputReconciler,
			"",
			t.GeneratedCodeRoot+"/input/reconciler.go",
			contrib.HybridSnapshotResources{
				LocalResourcesToSelect:  t.LocalInputResources,
				RemoteResourcesToSelect: t.RemoteInputResources,
			},
			snapshotApiGroups,
		))

		if t.AgentMode {
			topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
				contrib.InputReconciler,
				"Agent",
				t.GeneratedCodeRoot+"/input/agent_reconciler.go",
				contrib.HomogenousSnapshotResources{
					ResourcesToSelect: t.LocalInputResources.Join(t.RemoteInputResources),
				},
				snapshotApiGroups,
			))
		}
	case t.LocalInputResources != nil:
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputSnapshot,
			"",
			t.GeneratedCodeRoot+"/input/snapshot.go",
			contrib.HomogenousSnapshotResources{ResourcesToSelect: t.LocalInputResources},
			snapshotApiGroups,
		))
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputSnapshotManualBuilder,
			"",
			t.GeneratedCodeRoot+"/input/snapshot_manual_builder.go",
			contrib.HomogenousSnapshotResources{ResourcesToSelect: t.LocalInputResources},
			snapshotApiGroups,
		))
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputReconciler,
			"",
			t.GeneratedCodeRoot+"/input/reconciler.go",
			contrib.HomogenousSnapshotResources{ResourcesToSelect: t.LocalInputResources},
			snapshotApiGroups,
		))
	case t.RemoteInputResources != nil:
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputSnapshot,
			"",
			t.GeneratedCodeRoot+"/input/snapshot.go",
			contrib.HomogenousSnapshotResources{ResourcesToSelect: t.RemoteInputResources},
			snapshotApiGroups,
		))
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputSnapshotManualBuilder,
			"",
			t.GeneratedCodeRoot+"/input/snapshot_manual_builder.go",
			contrib.HomogenousSnapshotResources{ResourcesToSelect: t.RemoteInputResources},
			snapshotApiGroups,
		))
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputReconciler,
			"",
			t.GeneratedCodeRoot+"/input/reconciler.go",
			contrib.HomogenousSnapshotResources{ResourcesToSelect: t.RemoteInputResources},
			snapshotApiGroups,
		))
	}

	for _, outputResources := range t.OutputResources {
		filePath := filepath.Join(t.GeneratedCodeRoot, "output", outputResources.Name, "snapshot.go")
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.OutputSnapshot,
			"",
			filePath,
			contrib.HomogenousSnapshotResources{ResourcesToSelect: outputResources.Snapshot},
			snapshotApiGroups,
		))
	}

	return topLevelTemplates
}

func makeTopLevelTemplate(templateFunc func(params contrib.SnapshotTemplateParameters) model.CustomTemplates, snapshotName, outPath string, snapshotResources contrib.SnapshotResources, selectFromGroups map[string][]model.Group) model.CustomTemplates {
	return templateFunc(contrib.SnapshotTemplateParameters{
		SnapshotName:      snapshotName,
		OutputFilename:    outPath,
		SelectFromGroups:  selectFromGroups,
		SnapshotResources: snapshotResources,
	})
}
