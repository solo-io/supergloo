package main

/*
This file is used to import external api protos
*/

//go:generate go run import_protos.go

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/utils/log"
)

// all you should need to do is append to this!
var protosToImport = []importedProto{
	{
		crdGroupName: "networking.istio.io",
		file:         "https://raw.githubusercontent.com/istio/api/056eb85d96f09441775d79283c149d93fcbd0982/networking/v1alpha3/virtual_service.proto",
		importPath:   "github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3",
		skTypes: []soloKitType{
			{
				messageName: "VirtualService",
				shortName:   "virtualservice",
				pluralName:  "virtualservices",
			},
		},
	},
	{
		crdGroupName: "networking.istio.io",
		file:         "https://raw.githubusercontent.com/istio/api/056eb85d96f09441775d79283c149d93fcbd0982/networking/v1alpha3/destination_rule.proto",
		importPath:   "github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3",
		skTypes: []soloKitType{
			{
				messageName: "DestinationRule",
				shortName:   "destinationrule",
				pluralName:  "destinationrules",
			},
		},
	},
	{
		crdGroupName: "networking.istio.io",
		file:         "https://raw.githubusercontent.com/istio/api/056eb85d96f09441775d79283c149d93fcbd0982/networking/v1alpha3/sidecar.proto",
		importPath:   "github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3",
	},
	{
		crdGroupName: "networking.istio.io",
		file:         "https://raw.githubusercontent.com/istio/api/056eb85d96f09441775d79283c149d93fcbd0982/networking/v1alpha3/gateway.proto",
		importPath:   "github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3",
	},
	{
		crdGroupName: "authentication.istio.io",
		file:         "https://raw.githubusercontent.com/istio/api/056eb85d96f09441775d79283c149d93fcbd0982/authentication/v1alpha1/policy.proto",
		importPath:   "github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1",
		skTypes: []soloKitType{
			{
				messageName: "Policy",
				shortName:   "policy",
				pluralName:  "policies",
			},
			{
				messageName:   "MeshPolicy",
				shortName:     "meshpolicy",
				pluralName:    "meshpolicies",
				copyFrom:      "Policy",
				clusterScoped: true,
			},
		},
	},
	{
		crdGroupName: "rbac.istio.io",
		file:         "https://raw.githubusercontent.com/istio/api/056eb85d96f09441775d79283c149d93fcbd0982/rbac/v1alpha1/rbac.proto",
		importPath:   "github.com/solo-io/supergloo/pkg/api/external/istio/rbac/v1alpha1",
		skTypes: []soloKitType{
			{
				messageName: "ServiceRole",
				shortName:   "servicerole",
				pluralName:  "serviceroles",
			},
			{
				messageName: "ServiceRoleBinding",
				shortName:   "servicerolebinding",
				pluralName:  "servicerolebindings",
			},
			{
				messageName: "RbacConfig",
				shortName:   "rbacconfig",
				pluralName:  "rbacconfigs",
			},
		},
	},
}

func main() {
	for _, imp := range protosToImport {
		outDir := filepath.Join(os.Getenv("GOPATH"), "src", strings.Replace(imp.importPath, "pkg/", "", -1))
		os.MkdirAll(outDir, 0755)
		outFile := filepath.Join(outDir, filepath.Base(imp.file))
		out, err := os.Create(outFile)
		if err != nil {
			if os.IsNotExist(err) {
				out, err = os.Open(outFile)
				if err != nil {
					log.Fatalf("%v", err)
				}

			} else {
				log.Fatalf("%v", err)
			}
		}
		// create solo-kit.json for each imoportPath
		soloKitConfig, err := importIstioProto(imp.file, imp.importPath, imp.skTypes, out)
		if err != nil {
			log.Fatalf("%v", err)
		}
		if imp.crdGroupName != "" {
			soloKitConfig.CrdGroupOverride = imp.crdGroupName
		}
		b, err := json.MarshalIndent(soloKitConfig, "", "    ")
		if err != nil {
			log.Fatalf("%v", err)
		}
		if err := ioutil.WriteFile(filepath.Join(outDir, "solo-kit.json"), b, 0644); err != nil {
			log.Fatalf("%v", err)
		}
	}
}

func soloKitOptions(shortName, pluralName string, clusterScoped bool) string {
	opts := fmt.Sprintf(`
  option (core.solo.io.resource).short_name = "%v";
  option (core.solo.io.resource).plural_name = "%v";`, shortName, pluralName)
	if clusterScoped {
		opts += `
  option (core.solo.io.resource).cluster_scoped = true;`
	}
	return opts
}

const soloKitFields = `
  // Status indicates the validation status of this resource.
  // Status is read-only by clients, and set by supergloo during validation
  core.solo.io.Status status = 100 [(gogoproto.nullable) = false, (gogoproto.moretags) = "testdiff:\"ignore\""];

  // Metadata contains the object metadata for this resource
  core.solo.io.Metadata metadata = 101 [(gogoproto.nullable) = false];
`

const soloKitImports = `
import "github.com/solo-io/solo-kit/api/v1/metadata.proto";
import "github.com/solo-io/solo-kit/api/v1/status.proto";
import "github.com/solo-io/solo-kit/api/v1/solo-kit.proto";
import "gogoproto/gogo.proto";
option (gogoproto.equal_all) = true;
`

type importedProto struct {
	crdGroupName string
	file         string
	importPath   string
	skTypes      []soloKitType
}

type soloKitType struct {
	messageName   string
	shortName     string
	pluralName    string
	clusterScoped bool

	// we will clone the message definition from this type
	// required for Istio's MeshPolicy/Policy thing (same type, reuse the same proto)
	copyFrom string
}

func importIstioProto(file, importPath string, skTypes []soloKitType, out io.Writer) (*model.ProjectConfig, error) {
	resp, err := http.Get(file)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.Errorf("request failed with status code %v", resp.StatusCode)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	modifiedProto := replaceGoPackage(string(b), importPath)
	// duplicate messages first if necessary
	for _, skt := range skTypes {
		if skt.copyFrom != "" {
			modifiedProto = duplicateMessage(modifiedProto, skt.messageName, skt.copyFrom)
		}
	}
	for _, skt := range skTypes {
		modifiedProto = injectSoloKit(modifiedProto, skt.messageName, skt.shortName, skt.pluralName, skt.clusterScoped)
	}
	if _, err := fmt.Fprint(out, modifiedProto); err != nil {
		return nil, err
	}
	protoPackage, err := detectProtoPackage(modifiedProto)
	if err != nil {
		return nil, err
	}
	return &model.ProjectConfig{
		Name:    protoPackage,
		Version: filepath.Base(importPath),
	}, nil
}

var protoPackageStatementRegex = regexp.MustCompile(`\bpackage\b (.*);`)

func detectProtoPackage(protoContents string) (string, error) {
	matches := protoPackageStatementRegex.FindStringSubmatch(protoContents)
	if len(matches) != 2 {
		return "", errors.Errorf("invalid package statement: %v", matches)
	}
	return matches[1], nil
}

var goPackageStatementRegex = regexp.MustCompile(`option go_package.*=.*;`)

func replaceGoPackage(in, importPath string) string {
	in = strings.Replace(in, `import "gogoproto/gogo.proto";`, "", -1)
	in = strings.Replace(in, `option (gogoproto.equal_all) = true;`, "", -1)
	return goPackageStatementRegex.ReplaceAllString(in, fmt.Sprintf("option go_package = \"%v\";\n\n%v", importPath, soloKitImports))
}

func injectSoloKit(in, messageName, shortName, pluralName string, clusterScoped bool) string {
	messageDeclaration := regexp.MustCompile("message " + messageName + " {")
	updatedMessageDeclaration := fmt.Sprintf("message %v {\n"+
		"%v\n"+
		"%v\n",
		messageName, soloKitOptions(shortName, pluralName, clusterScoped), soloKitFields)
	return messageDeclaration.ReplaceAllString(in, updatedMessageDeclaration)
}

func duplicateMessage(in, destMessage, sourceMessage string) string {
	entireMessageDefinition := regexp.MustCompile(`(?s)message ` + sourceMessage + ` \{.+?\n\}`)
	duplicate := entireMessageDefinition.FindString(in)
	renamed := "// " + destMessage + " copied from " + sourceMessage + "\n" + strings.Replace(duplicate, sourceMessage, destMessage, -1)
	in = entireMessageDefinition.ReplaceAllString(in,
		duplicate+
			"\n\n"+
			renamed)
	return in
}
