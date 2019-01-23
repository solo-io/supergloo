package utils

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	. "github.com/onsi/gomega"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/v1"
)

func SetupAppMesh(meshClient v1.MeshClient, secretClient gloov1.SecretClient, namespace string) core.ResourceRef {
	secret, mesh := MakeAppMeshResources(namespace)
	secretClient.Delete(namespace, secret.Metadata.Name, clients.DeleteOpts{})
	_, err := secretClient.Write(secret, clients.WriteOpts{})
	Expect(err).NotTo(HaveOccurred())

	meshClient.Delete(namespace, mesh.Metadata.Name, clients.DeleteOpts{})

	mesh1, err := meshClient.Write(mesh, clients.WriteOpts{})
	Expect(err).NotTo(HaveOccurred())
	Expect(mesh1).NotTo(BeNil())
	return mesh1.Metadata.Ref()
}

func MakeAppMeshResources(namespace string) (*gloov1.Secret, *v1.Mesh) {
	creds, err := credentials.NewSharedCredentials("", "").Get()
	Expect(err).NotTo(HaveOccurred())
	secretMeta := core.Metadata{Name: "my-appmesh-credentials", Namespace: namespace}
	meshMeta := core.Metadata{Name: "my-appmesh", Namespace: namespace}
	secretRef := secretMeta.Ref()
	return &gloov1.Secret{
			Metadata: secretMeta,
			Kind: &gloov1.Secret_Aws{
				Aws: &gloov1.AwsSecret{
					// these can be read in from ~/.aws/credentials by default (if user does not provide)
					// see https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html for more details
					AccessKey: creds.AccessKeyID,
					SecretKey: creds.SecretAccessKey,
				},
			},
		}, &v1.Mesh{
			Metadata: meshMeta,
			MeshType: &v1.Mesh_AppMesh{
				AppMesh: &v1.AppMesh{
					AwsRegion:      "us-east-1",
					AwsCredentials: &secretRef,
				},
			},
		}
}
