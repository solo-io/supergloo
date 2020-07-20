package dockerutils

import (
	"github.com/docker/distribution/reference"
)

type Image struct {
	Domain string `json:"domain,omitempty"`
	Path   string `json:"path,omitempty"`
	Tag    string `json:"tag,omitempty"`
	Digest string `json:"digest,omitempty"`
}

func ParseImageName(image string) (*Image, error) {
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
