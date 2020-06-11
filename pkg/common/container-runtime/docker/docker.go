package docker

import (
	"github.com/docker/distribution/reference"
)

type Image struct {
	Domain string `json:"domain,omitempty"`
	Path   string `json:"path,omitempty"`
	Tag    string `json:"tag,omitempty"`
	Digest string `json:"digest,omitempty"`
}

//go:generate mockgen -source docker.go -destination mocks/docker.go
type ImageNameParser interface {
	// Parses a tagged or digested docker image name into its different parts
	// The image's path is normalized (i.e., calling this function on "consul:1.6.2" returns a path of "library/consul")
	Parse(image string) (*Image, error)
}

type imageNameParser struct{}

func NewImageNameParser() ImageNameParser {
	return &imageNameParser{}
}

func (i *imageNameParser) Parse(image string) (*Image, error) {
	named, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return nil, err
	}

	tagged, isTagged := named.(reference.Tagged)
	digested, isDigested := named.(reference.Digested)

	var tag, digest string
	if isTagged {
		tag = tagged.Tag()
	} else if isDigested {
		digest = digested.Digest().String()
	}

	return &Image{
		Domain: reference.Domain(named),
		Path:   reference.Path(named),
		Tag:    tag,
		Digest: digest,
	}, nil
}
