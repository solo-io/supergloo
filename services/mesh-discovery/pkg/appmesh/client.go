package appmesh

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/pkg/errors"
)

// Represents the App Mesh API
type Client interface {
	// Get operations
	GetMesh(ctx context.Context, meshName string) (*appmesh.MeshData, error)

	// List operations
	ListMeshes(ctx context.Context) ([]string, error)
}

type client struct {
	api *appmesh.AppMesh
}

func (c *client) GetMesh(ctx context.Context, meshName string) (*appmesh.MeshData, error) {
	input := &appmesh.DescribeMeshInput{
		MeshName: aws.String(meshName),
	}
	if output, err := c.api.DescribeMeshWithContext(ctx, input); err != nil {
		if IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "failed to describe mesh %s", meshName)
	} else if output == nil {
		return nil, nil
	} else {
		return output.Mesh, nil
	}
}

func (c *client) ListMeshes(ctx context.Context) ([]string, error) {
	var meshRefs []*appmesh.MeshRef

	paginationFunc := func(output *appmesh.ListMeshesOutput, b bool) bool {
		meshRefs = append(meshRefs, output.Meshes...)
		return true
	}

	if err := c.api.ListMeshesPagesWithContext(ctx, &appmesh.ListMeshesInput{}, paginationFunc); err != nil {
		return nil, errors.Wrapf(err, "failed to list meshes")
	}

	var result []string
	for _, meshRef := range meshRefs {
		result = append(result, *meshRef.MeshName)
	}

	return result, nil
}

func IsNotFound(err error) bool {
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == appmesh.ErrCodeNotFoundException {
				return true
			}
		}
	}
	return false
}
