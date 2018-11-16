// Code generated by protoc-gen-solo-kit. DO NOT EDIT.

package v1

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
)

type InstallClient interface {
	BaseClient() clients.ResourceClient
	Register() error
	Read(namespace, name string, opts clients.ReadOpts) (*Install, error)
	Write(resource *Install, opts clients.WriteOpts) (*Install, error)
	Delete(namespace, name string, opts clients.DeleteOpts) error
	List(namespace string, opts clients.ListOpts) (InstallList, error)
	Watch(namespace string, opts clients.WatchOpts) (<-chan InstallList, <-chan error, error)
}

type installClient struct {
	rc clients.ResourceClient
}

func NewInstallClient(rcFactory factory.ResourceClientFactory) (InstallClient, error) {
	return NewInstallClientWithToken(rcFactory, "")
}

func NewInstallClientWithToken(rcFactory factory.ResourceClientFactory, token string) (InstallClient, error) {
	rc, err := rcFactory.NewResourceClient(factory.NewResourceClientParams{
		ResourceType: &Install{},
		Token:        token,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating base Install resource client")
	}
	return &installClient{
		rc: rc,
	}, nil
}

func (client *installClient) BaseClient() clients.ResourceClient {
	return client.rc
}

func (client *installClient) Register() error {
	return client.rc.Register()
}

func (client *installClient) Read(namespace, name string, opts clients.ReadOpts) (*Install, error) {
	opts = opts.WithDefaults()
	resource, err := client.rc.Read(namespace, name, opts)
	if err != nil {
		return nil, err
	}
	return resource.(*Install), nil
}

func (client *installClient) Write(install *Install, opts clients.WriteOpts) (*Install, error) {
	opts = opts.WithDefaults()
	resource, err := client.rc.Write(install, opts)
	if err != nil {
		return nil, err
	}
	return resource.(*Install), nil
}

func (client *installClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()
	return client.rc.Delete(namespace, name, opts)
}

func (client *installClient) List(namespace string, opts clients.ListOpts) (InstallList, error) {
	opts = opts.WithDefaults()
	resourceList, err := client.rc.List(namespace, opts)
	if err != nil {
		return nil, err
	}
	return convertToInstall(resourceList), nil
}

func (client *installClient) Watch(namespace string, opts clients.WatchOpts) (<-chan InstallList, <-chan error, error) {
	opts = opts.WithDefaults()
	resourcesChan, errs, initErr := client.rc.Watch(namespace, opts)
	if initErr != nil {
		return nil, nil, initErr
	}
	installsChan := make(chan InstallList)
	go func() {
		for {
			select {
			case resourceList := <-resourcesChan:
				installsChan <- convertToInstall(resourceList)
			case <-opts.Ctx.Done():
				close(installsChan)
				return
			}
		}
	}()
	return installsChan, errs, nil
}

func convertToInstall(resources resources.ResourceList) InstallList {
	var installList InstallList
	for _, resource := range resources {
		installList = append(installList, resource.(*Install))
	}
	return installList
}
