package model

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/plugin"
	"github.com/ilackarms/protokit"
	"github.com/solo-io/solo-kit/pkg/errors"
)

const ProjectConfigFilename = "solo-kit.json"

// SOLO-KIT Descriptors from which code can be generated

type ProjectConfig struct {
	Title          string                      `json:"title"`
	Description    string                      `json:"description"`
	Name           string                      `json:"name"`
	Version        string                      `json:"version"`
	DocsDir        string                      `json:"docs_dir"`
	ResourceGroups map[string][]ResourceConfig `json:"resource_groups"`
	// set by load
	ProjectFile string
	GoPackage   string
}

type ResourceConfig struct {
	MessageName    string `json:"name"`
	MessagePackage string `json:"package"`
}

type Project struct {
	ProjectConfig  ProjectConfig
	ProtoPackage   string
	Resources      []*Resource
	ResourceGroups []*ResourceGroup
	XDSResources   []*XDSResource

	Request *plugin_go.CodeGeneratorRequest
}

type Resource struct {
	Name         string
	PluralName   string
	ShortName    string
	ProtoPackage string // eg. gloo.solo.io
	// ImportPrefix will equal ProtoPackage+"." if the resource does not belong to the project
	// else it will be empty string. used in event loop files
	ImportPrefix string
	// empty unless resource is external
	// format "github.com/solo-io/solo-kit/foo/bar"
	GoPackage string

	HasStatus     bool
	ClusterScoped bool // the resource lives at the cluster level, namespace is ignored

	Fields []*Field
	Oneofs []*Oneof

	// resource groups i belong to
	ResourceGroups []*ResourceGroup
	// project i belong to
	Project *Project

	Filename string // the proto file where this resource is contained

	// TODO (ilackarms): change to use descriptor.DescriptorProto
	Original *protokit.Descriptor
}

type Field struct {
	Name        string
	TypeName    string
	IsOneof     bool // we ignore oneof fields in test generation
	SkipHashing bool // skip this field when hashing the resource
	Original    *descriptor.FieldDescriptorProto
}

type Oneof struct {
	Name string
}

type ResourceGroup struct {
	Name      string // eg. api.gloo.solo.io
	GoName    string // will be Api
	Imports   string // if this resource group contains any imports from other projects
	Project   *Project
	Resources []*Resource
}

type XDSResource struct {
	Name         string
	MessageType  string
	NameField    string
	NoReferences bool

	Project      *Project
	ProtoPackage string // eg. gloo.solo.io
}

func LoadProjectConfig(path string) (ProjectConfig, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return ProjectConfig{}, err
	}
	var pc ProjectConfig
	err = json.Unmarshal(b, &pc)
	pc.ProjectFile = path
	goPkg, err := detectGoPackageForProject(path)
	if err != nil {
		return ProjectConfig{}, err
	}
	pc.GoPackage = goPkg
	return pc, err
}

var goPackageStatementRegex = regexp.MustCompile(`option go_package = "(.*)";`)

func detectGoPackageForProject(projectFile string) (string, error) {
	var goPkg string
	projectDir := filepath.Dir(projectFile)
	if err := filepath.Walk(projectDir, func(protoFile string, info os.FileInfo, err error) error {
		// already set
		if goPkg != "" {
			return nil
		}
		if !strings.HasSuffix(protoFile, ".proto") {
			return nil
		}
		// search for go_package on protos in the same dir as the project.json
		if projectDir != filepath.Dir(protoFile) {
			return nil
		}
		content, err := ioutil.ReadFile(protoFile)
		if err != nil {
			return err
		}
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			goPackage := goPackageStatementRegex.FindStringSubmatch(line)
			if len(goPackage) == 0 {
				continue
			}
			if len(goPackage) != 2 {
				return errors.Errorf("parsing go_package error: from %v found %v", line, goPackage)
			}
			goPkg = goPackage[1]
			break
		}
		return nil
	}); err != nil {
		return "", err
	}
	if goPkg == "" {
		return "", errors.Errorf("no go_package statement found in root dir of project %v", projectFile)
	}
	return goPkg, nil
}
