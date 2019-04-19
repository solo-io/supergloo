// Code generated by solo-kit. DO NOT EDIT.

package v1

import (
	"sync"
	"time"

	gloo_solo_io "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	istio_authentication_v1alpha1 "github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/errutils"
)

var (
	mIstioDiscoverySnapshotIn  = stats.Int64("istioDiscovery.supergloo.solo.io/snap_emitter/snap_in", "The number of snapshots in", "1")
	mIstioDiscoverySnapshotOut = stats.Int64("istioDiscovery.supergloo.solo.io/snap_emitter/snap_out", "The number of snapshots out", "1")

	istioDiscoverysnapshotInView = &view.View{
		Name:        "istioDiscovery.supergloo.solo.io_snap_emitter/snap_in",
		Measure:     mIstioDiscoverySnapshotIn,
		Description: "The number of snapshots updates coming in",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{},
	}
	istioDiscoverysnapshotOutView = &view.View{
		Name:        "istioDiscovery.supergloo.solo.io/snap_emitter/snap_out",
		Measure:     mIstioDiscoverySnapshotOut,
		Description: "The number of snapshots updates going out",
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{},
	}
)

func init() {
	view.Register(istioDiscoverysnapshotInView, istioDiscoverysnapshotOutView)
}

type IstioDiscoveryEmitter interface {
	Register() error
	Mesh() MeshClient
	Install() InstallClient
	KubeNamespace() KubeNamespaceClient
	Pod() PodClient
	Upstream() gloo_solo_io.UpstreamClient
	MeshPolicy() istio_authentication_v1alpha1.MeshPolicyClient
	Snapshots(watchNamespaces []string, opts clients.WatchOpts) (<-chan *IstioDiscoverySnapshot, <-chan error, error)
}

func NewIstioDiscoveryEmitter(meshClient MeshClient, installClient InstallClient, kubeNamespaceClient KubeNamespaceClient, podClient PodClient, upstreamClient gloo_solo_io.UpstreamClient, meshPolicyClient istio_authentication_v1alpha1.MeshPolicyClient) IstioDiscoveryEmitter {
	return NewIstioDiscoveryEmitterWithEmit(meshClient, installClient, kubeNamespaceClient, podClient, upstreamClient, meshPolicyClient, make(chan struct{}))
}

func NewIstioDiscoveryEmitterWithEmit(meshClient MeshClient, installClient InstallClient, kubeNamespaceClient KubeNamespaceClient, podClient PodClient, upstreamClient gloo_solo_io.UpstreamClient, meshPolicyClient istio_authentication_v1alpha1.MeshPolicyClient, emit <-chan struct{}) IstioDiscoveryEmitter {
	return &istioDiscoveryEmitter{
		mesh:          meshClient,
		install:       installClient,
		kubeNamespace: kubeNamespaceClient,
		pod:           podClient,
		upstream:      upstreamClient,
		meshPolicy:    meshPolicyClient,
		forceEmit:     emit,
	}
}

type istioDiscoveryEmitter struct {
	forceEmit     <-chan struct{}
	mesh          MeshClient
	install       InstallClient
	kubeNamespace KubeNamespaceClient
	pod           PodClient
	upstream      gloo_solo_io.UpstreamClient
	meshPolicy    istio_authentication_v1alpha1.MeshPolicyClient
}

func (c *istioDiscoveryEmitter) Register() error {
	if err := c.mesh.Register(); err != nil {
		return err
	}
	if err := c.install.Register(); err != nil {
		return err
	}
	if err := c.kubeNamespace.Register(); err != nil {
		return err
	}
	if err := c.pod.Register(); err != nil {
		return err
	}
	if err := c.upstream.Register(); err != nil {
		return err
	}
	if err := c.meshPolicy.Register(); err != nil {
		return err
	}
	return nil
}

func (c *istioDiscoveryEmitter) Mesh() MeshClient {
	return c.mesh
}

func (c *istioDiscoveryEmitter) Install() InstallClient {
	return c.install
}

func (c *istioDiscoveryEmitter) KubeNamespace() KubeNamespaceClient {
	return c.kubeNamespace
}

func (c *istioDiscoveryEmitter) Pod() PodClient {
	return c.pod
}

func (c *istioDiscoveryEmitter) Upstream() gloo_solo_io.UpstreamClient {
	return c.upstream
}

func (c *istioDiscoveryEmitter) MeshPolicy() istio_authentication_v1alpha1.MeshPolicyClient {
	return c.meshPolicy
}

func (c *istioDiscoveryEmitter) Snapshots(watchNamespaces []string, opts clients.WatchOpts) (<-chan *IstioDiscoverySnapshot, <-chan error, error) {

	if len(watchNamespaces) == 0 {
		watchNamespaces = []string{""}
	}

	for _, ns := range watchNamespaces {
		if ns == "" && len(watchNamespaces) > 1 {
			return nil, nil, errors.Errorf("the \"\" namespace is used to watch all namespaces. Snapshots can either be tracked for " +
				"specific namespaces or \"\" AllNamespaces, but not both.")
		}
	}

	errs := make(chan error)
	var done sync.WaitGroup
	ctx := opts.Ctx
	/* Create channel for Mesh */
	type meshListWithNamespace struct {
		list      MeshList
		namespace string
	}
	meshChan := make(chan meshListWithNamespace)
	/* Create channel for Install */
	type installListWithNamespace struct {
		list      InstallList
		namespace string
	}
	installChan := make(chan installListWithNamespace)
	/* Create channel for KubeNamespace */
	type kubeNamespaceListWithNamespace struct {
		list      KubeNamespaceList
		namespace string
	}
	kubeNamespaceChan := make(chan kubeNamespaceListWithNamespace)
	/* Create channel for Pod */
	type podListWithNamespace struct {
		list      PodList
		namespace string
	}
	podChan := make(chan podListWithNamespace)
	/* Create channel for Upstream */
	type upstreamListWithNamespace struct {
		list      gloo_solo_io.UpstreamList
		namespace string
	}
	upstreamChan := make(chan upstreamListWithNamespace)
	/* Create channel for MeshPolicy */

	for _, namespace := range watchNamespaces {
		/* Setup namespaced watch for Mesh */
		meshNamespacesChan, meshErrs, err := c.mesh.Watch(namespace, opts)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "starting Mesh watch")
		}

		done.Add(1)
		go func(namespace string) {
			defer done.Done()
			errutils.AggregateErrs(ctx, errs, meshErrs, namespace+"-meshes")
		}(namespace)
		/* Setup namespaced watch for Install */
		installNamespacesChan, installErrs, err := c.install.Watch(namespace, opts)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "starting Install watch")
		}

		done.Add(1)
		go func(namespace string) {
			defer done.Done()
			errutils.AggregateErrs(ctx, errs, installErrs, namespace+"-installs")
		}(namespace)
		/* Setup namespaced watch for KubeNamespace */
		kubeNamespaceNamespacesChan, kubeNamespaceErrs, err := c.kubeNamespace.Watch(namespace, opts)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "starting KubeNamespace watch")
		}

		done.Add(1)
		go func(namespace string) {
			defer done.Done()
			errutils.AggregateErrs(ctx, errs, kubeNamespaceErrs, namespace+"-kubenamespaces")
		}(namespace)
		/* Setup namespaced watch for Pod */
		podNamespacesChan, podErrs, err := c.pod.Watch(namespace, opts)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "starting Pod watch")
		}

		done.Add(1)
		go func(namespace string) {
			defer done.Done()
			errutils.AggregateErrs(ctx, errs, podErrs, namespace+"-pods")
		}(namespace)
		/* Setup namespaced watch for Upstream */
		upstreamNamespacesChan, upstreamErrs, err := c.upstream.Watch(namespace, opts)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "starting Upstream watch")
		}

		done.Add(1)
		go func(namespace string) {
			defer done.Done()
			errutils.AggregateErrs(ctx, errs, upstreamErrs, namespace+"-upstreams")
		}(namespace)

		/* Watch for changes and update snapshot */
		go func(namespace string) {
			for {
				select {
				case <-ctx.Done():
					return
				case meshList := <-meshNamespacesChan:
					select {
					case <-ctx.Done():
						return
					case meshChan <- meshListWithNamespace{list: meshList, namespace: namespace}:
					}
				case installList := <-installNamespacesChan:
					select {
					case <-ctx.Done():
						return
					case installChan <- installListWithNamespace{list: installList, namespace: namespace}:
					}
				case kubeNamespaceList := <-kubeNamespaceNamespacesChan:
					select {
					case <-ctx.Done():
						return
					case kubeNamespaceChan <- kubeNamespaceListWithNamespace{list: kubeNamespaceList, namespace: namespace}:
					}
				case podList := <-podNamespacesChan:
					select {
					case <-ctx.Done():
						return
					case podChan <- podListWithNamespace{list: podList, namespace: namespace}:
					}
				case upstreamList := <-upstreamNamespacesChan:
					select {
					case <-ctx.Done():
						return
					case upstreamChan <- upstreamListWithNamespace{list: upstreamList, namespace: namespace}:
					}
				}
			}
		}(namespace)
	}
	/* Setup cluster-wide watch for MeshPolicy */

	meshPolicyChan, meshPolicyErrs, err := c.meshPolicy.Watch(opts)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "starting MeshPolicy watch")
	}
	done.Add(1)
	go func() {
		defer done.Done()
		errutils.AggregateErrs(ctx, errs, meshPolicyErrs, "meshpolicies")
	}()

	snapshots := make(chan *IstioDiscoverySnapshot)
	go func() {
		originalSnapshot := IstioDiscoverySnapshot{}
		currentSnapshot := originalSnapshot.Clone()
		timer := time.NewTicker(time.Second * 1)
		sync := func() {
			if originalSnapshot.Hash() == currentSnapshot.Hash() {
				return
			}

			stats.Record(ctx, mIstioDiscoverySnapshotOut.M(1))
			originalSnapshot = currentSnapshot.Clone()
			sentSnapshot := currentSnapshot.Clone()
			snapshots <- &sentSnapshot
		}

		for {
			record := func() { stats.Record(ctx, mIstioDiscoverySnapshotIn.M(1)) }

			select {
			case <-timer.C:
				sync()
			case <-ctx.Done():
				close(snapshots)
				done.Wait()
				close(errs)
				return
			case <-c.forceEmit:
				sentSnapshot := currentSnapshot.Clone()
				snapshots <- &sentSnapshot
			case meshNamespacedList := <-meshChan:
				record()

				namespace := meshNamespacedList.namespace
				meshList := meshNamespacedList.list

				currentSnapshot.Meshes[namespace] = meshList
			case installNamespacedList := <-installChan:
				record()

				namespace := installNamespacedList.namespace
				installList := installNamespacedList.list

				currentSnapshot.Installs[namespace] = installList
			case kubeNamespaceNamespacedList := <-kubeNamespaceChan:
				record()

				namespace := kubeNamespaceNamespacedList.namespace
				kubeNamespaceList := kubeNamespaceNamespacedList.list

				currentSnapshot.Kubenamespaces[namespace] = kubeNamespaceList
			case podNamespacedList := <-podChan:
				record()

				namespace := podNamespacedList.namespace
				podList := podNamespacedList.list

				currentSnapshot.Pods[namespace] = podList
			case upstreamNamespacedList := <-upstreamChan:
				record()

				namespace := upstreamNamespacedList.namespace
				upstreamList := upstreamNamespacedList.list

				currentSnapshot.Upstreams[namespace] = upstreamList
			case meshPolicyList := <-meshPolicyChan:
				record()
				currentSnapshot.Meshpolicies = meshPolicyList
			}
		}
	}()
	return snapshots, errs, nil
}
