package funcs

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"unicode"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/iancoleman/strcase"
	"github.com/ilackarms/protoc-gen-doc"
	"github.com/ilackarms/protokit"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/errors"
)

var primitiveTypes = map[descriptor.FieldDescriptorProto_Type]string{
	descriptor.FieldDescriptorProto_TYPE_FLOAT:  "float",
	descriptor.FieldDescriptorProto_TYPE_DOUBLE: "float",
	descriptor.FieldDescriptorProto_TYPE_STRING: "string",
	descriptor.FieldDescriptorProto_TYPE_BOOL:   "bool",
	descriptor.FieldDescriptorProto_TYPE_UINT32: "int",
	descriptor.FieldDescriptorProto_TYPE_UINT64: "int",
	descriptor.FieldDescriptorProto_TYPE_INT32:  "int",
	descriptor.FieldDescriptorProto_TYPE_INT64:  "int",
	descriptor.FieldDescriptorProto_TYPE_BYTES:  "bytes",
}

// container for the funcmap, allows functions to utilize each other more easily
type templateFunctions struct {
	Funcs template.FuncMap
}

var magicCommentRegex = regexp.MustCompile("@solo-kit:.*")
var githubProjectFileRegex = regexp.MustCompile(".*github.com/([^/]*)/([^/]*)/(.*)")

func TemplateFuncs(project *model.Project) template.FuncMap {
	funcs := &templateFunctions{}
	funcMap := template.FuncMap{
		"join":               strings.Join,
		"lowercase":          strings.ToLower,
		"lower_camel":        strcase.ToLowerCamel,
		"upper_camel":        strcase.ToCamel,
		"snake":              strcase.ToSnake,
		"p":                  gendoc.PFilter,
		"para":               gendoc.ParaFilter,
		"nobr":               gendoc.NoBrFilter,
		"fieldType":          fieldType(project),
		"yamlType":           yamlType,
		"noescape":           noEscape,
		"linkForField":       linkForField(project),
		"linkForResource":    linkForResource(project),
		"forEachMessage":     funcs.forEachMessage,
		"resourceForMessage": resourceForMessage(project),
		"getFileForMessage": func(msg *protokit.Descriptor) *protokit.FileDescriptor {
			return msg.GetFile()
		},
		// assumes the file lives in a github-hosted repo
		"githubLinkForFile": func(branch, path string) (string, error) {
			githubFile := githubProjectFileRegex.FindStringSubmatch(path)
			if len(githubFile) != 4 {
				return "`" + path + "`", nil
				//return "", errors.Errorf("invalid path provided, must contain github.com/ in path: %v", path)
			}
			org := githubFile[1]
			project := githubFile[2]
			suffix := githubFile[3]
			return fmt.Sprintf("[%v](https://github.com/%v/%v/blob/%v/%v)",
				path, org, project, branch, suffix), nil
		},
		"printfptr": printPointer,
		"remove_magic_comments": func(in string) string {
			lines := strings.Split(in, "\n")
			var linesWithoutMagicComments []string
			for _, line := range lines {
				if magicCommentRegex.MatchString(line) {
					continue
				}
				linesWithoutMagicComments = append(linesWithoutMagicComments, line)
			}
			return strings.Join(linesWithoutMagicComments, "\n")
		},
		"new_str_slice": func() *[]string {
			var v []string
			return &v
		},
		"append_str_slice": func(to *[]string, str string) *[]string {
			*to = append(*to, str)
			return to
		},
		"join_str_slice": func(slc *[]string, sep string) string {
			return strings.Join(*slc, sep)
		},
		"new_bool": func() *bool {
			var v bool
			return &v
		},
		"set_bool": func(v *bool, val bool) *bool {
			*v = val
			return v
		},
		"backtick": func() string {
			return "`"
		},
	}
	funcs.Funcs = funcMap
	return funcMap
}

func printPointer(format string, p *string) string {
	return fmt.Sprintf(format, *p)
}

func yamlType(longType, label string) string {
	yamlType := func() string {
		if strings.HasPrefix(longType, "map<") {
			return longType
		}
		switch longType {
		case "string":
			fallthrough
		case "uint32":
			fallthrough
		case "bool":
			fallthrough
		case "int32":
			return longType
		case "Status":
			return "(read only)"
		}
		return "{" + longType + "}"
	}()
	if label == "repeated" {
		yamlType = "[" + yamlType + "]"
	}
	return yamlType
}

func noEscape(s string) htmltemplate.HTML {
	return htmltemplate.HTML(s)
}

func fieldType(project *model.Project) func(field *protokit.FieldDescriptor) (string, error) {
	return func(field *protokit.FieldDescriptor) (string, error) {
		fieldTypeStr := func() string {
			switch field.GetType() {
			case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
				return field.GetTypeName()
			case descriptor.FieldDescriptorProto_TYPE_ENUM:
				return field.GetTypeName()
			}
			if typeName, ok := primitiveTypes[field.GetType()]; ok {
				return typeName
			}
			return "UNSUPPORTED: " + field.GetType().String() + ": " + field.GetName()
		}()
		if field.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			fieldTypeStr = "[]" + strings.TrimPrefix(fieldTypeStr, ".")
		}
		if strings.HasSuffix(fieldTypeStr, "Entry") {
			_, msg, enum, err := getFileAndTypeDefForField(project, field)
			if err != nil {
				return "", err
			}
			if enum != nil {
				return "", errors.Errorf("unexpected enum %v for field type %v", enum.GetName(), fieldTypeStr)
			}
			if len(msg.Field) != 2 {
				return "", errors.Errorf("message %v was Entry type, expected map", msg.GetName())
			}
			key, err := fieldType(project)(&protokit.FieldDescriptor{FieldDescriptorProto: msg.Field[0]})
			if err != nil {
				return "", err
			}
			val, err := fieldType(project)(&protokit.FieldDescriptor{FieldDescriptorProto: msg.Field[1]})
			if err != nil {
				return "", err
			}
			fieldTypeStr = "map<" + key + ", " + val + ">"
		}
		return fieldTypeStr, nil
	}
}

func wellKnownProtoLink(typeName string) string {
	wellKnown := strings.TrimPrefix(typeName, ".google.protobuf.")
	wellKnown = strcase.ToSnake(wellKnown)
	wellKnown = strings.Replace(wellKnown, "_", "-", -1)
	wellKnown = "https://developers.google.com/protocol-buffers/docs/reference/csharp/class/google/protobuf/well-known-types/" + wellKnown
	return wellKnown
}

func linkForField(project *model.Project) func(forFile *protokit.FileDescriptor, field *protokit.FieldDescriptor) (string, error) {
	return func(forFile *protokit.FileDescriptor, field *protokit.FieldDescriptor) (string, error) {
		typeName, err := fieldType(project)(field)
		if err != nil {
			return "", err
		}
		if _, ok := primitiveTypes[field.GetType()]; ok || strings.HasPrefix(typeName, "map<") {
			return "`" + typeName + "`", nil
		}
		file, msg, enum, err := getFileAndTypeDefForField(project, field)
		if err != nil {
			return "", err
		}
		var declaredName string
		if msg != nil {
			declaredName = msg.GetName()
		} else {
			declaredName = enum.GetName()
		}
		var link string
		switch {
		case strings.Contains(typeName, ".google.protobuf."):
			link = wellKnownProtoLink(typeName)
		default:
			var linkedFile string
			for _, toGenerate := range project.Request.FileToGenerate {
				if strings.HasSuffix(file.GetName(), toGenerate) {
					linkedFile = toGenerate
					break
				}
			}
			if linkedFile == "" {
				linkedFile = filepath.Base(file.GetName())
				//return "", errors.Errorf("failed to get generated file path for proto %v in list %v", file.GetName(), project.Request.FileToGenerate)
			}
			linkedFile = relativeFilename(forFile.GetName(), linkedFile)
			link = linkedFile + ".sk.md#" + declaredName
		}
		linkText := "[" + typeName + "](" + link + ")"
		return linkText, nil
	}
}

func linkForResource(project *model.Project) func(resource *model.Resource) (string, error) {
	protoFiles := protokit.ParseCodeGenRequest(project.Request)
	return func(resource *model.Resource) (string, error) {
		for _, file := range protoFiles {
			if file.GetName() == resource.Filename {
				// TODO: turn this X.proto.sk.md convention into a function lest this linking break
				return fmt.Sprintf("[%v](./%v.sk.md#%v)", resource.Name, resource.Filename, resource.Name), nil
			}
		}
		return "", errors.Errorf("internal error: could not find file for resource %v in project %v",
			resource.Filename, project.ProjectConfig.Name)
	}
}

func resourceForMessage(project *model.Project) func(msg *protokit.Descriptor) (*model.Resource, error) {
	return func(msg *protokit.Descriptor) (*model.Resource, error) {
		for _, res := range project.Resources {
			if res.Original.GetName() == msg.GetName() && res.Original.GetFile().GetName() == msg.GetFile().GetName() {
				return res, nil
			}
		}
		return nil, nil
		return nil, errors.Errorf("internal error: could not find file for resource for msg %v in project %v",
			msg.GetName(), project.ProjectConfig.Name)
	}
}

func relativeFilename(fileWithLink, fileLinkedTo string) string {
	if fileLinkedTo == fileWithLink {
		return filepath.Base(fileLinkedTo)
	}
	prefix := commonPrefix(os.PathSeparator, fileWithLink, fileLinkedTo) + string(os.PathSeparator)
	fileWithLink = strings.TrimPrefix(fileWithLink, prefix)
	fileLinkedTo = strings.TrimPrefix(fileLinkedTo, prefix)
	fileWithLinkSplit := strings.Split(fileWithLink, string(os.PathSeparator))
	if len(fileWithLinkSplit) == 1 {
		return fileLinkedTo
	}
	for i := 0; i < len(fileWithLinkSplit)-1; i++ {
		fileLinkedTo = ".." + string(os.PathSeparator) + fileLinkedTo
	}
	return fileLinkedTo
}

// from https://www.rosettacode.org/wiki/Find_common_directory_path#Go
func commonPrefix(sep byte, paths ...string) string {
	// Handle special cases.
	switch len(paths) {
	case 0:
		return ""
	case 1:
		return path.Clean(paths[0])
	}

	// Note, we treat string as []byte, not []rune as is often
	// done in Go. (And sep as byte, not rune). This is because
	// most/all supported OS' treat paths as string of non-zero
	// bytes. A filename may be displayed as a sequence of Unicode
	// runes (typically encoded as UTF-8) but paths are
	// not required to be valid UTF-8 or in any normalized form
	// (e.g. "é" (U+00C9) and "é" (U+0065,U+0301) are different
	// file names.
	c := []byte(path.Clean(paths[0]))

	// We add a trailing sep to handle the case where the
	// common prefix directory is included in the path list
	// (e.g. /home/user1, /home/user1/foo, /home/user1/bar).
	// path.Clean will have cleaned off trailing / separators with
	// the exception of the root directory, "/" (in which case we
	// make it "//", but this will get fixed up to "/" bellow).
	c = append(c, sep)

	// Ignore the first path since it's already in c
	for _, v := range paths[1:] {
		// Clean up each path before testing it
		v = path.Clean(v) + string(sep)

		// Find the first non-common byte and truncate c
		if len(v) < len(c) {
			c = c[:len(v)]
		}
		for i := 0; i < len(c); i++ {
			if v[i] != c[i] {
				c = c[:i]
				break
			}
		}
	}

	// Remove trailing non-separator characters and the final separator
	for i := len(c) - 1; i >= 0; i-- {
		if c[i] == sep {
			c = c[:i]
			break
		}
	}

	return string(c)
}

func getFileForField(project *model.Project, field *protokit.FieldDescriptor) (*descriptor.FileDescriptorProto, error) {
	parts := strings.Split(strings.TrimPrefix(field.GetTypeName(), "."), ".")
	if strings.HasSuffix(parts[len(parts)-1], "Entry") {
		parts = parts[:len(parts)-1]
	}
	messageName := parts[len(parts)-1]
	packageName := strings.Join(parts[:len(parts)-1], ".")
	for _, protoFile := range project.Request.GetProtoFile() {
		if protoFile.GetPackage() == packageName {
			for _, msg := range protoFile.GetMessageType() {
				if messageName == msg.GetName() {
					return protoFile, nil
				}
			}
		}
	}
	for _, protoFile := range project.Request.ProtoFile {
		// ilackarms: unlikely event of collision where the package name has the right prefix and a nested message type matches
		if strings.HasPrefix(packageName, protoFile.GetPackage()) {
			for _, msg := range protoFile.GetMessageType() {
				for _, nestedMsg := range msg.GetNestedType() {
					if messageName == nestedMsg.GetName() {
						return protoFile, nil
					}
				}
			}
		}
	}
	return nil, errors.Errorf("message %v.%v not found", packageName, messageName)
}

func splitTypeName(typeName string) (string, []string) {
	parts := strings.Split(strings.TrimPrefix(typeName, "."), ".")
	var indexOfFirstUppercasePart int
	for i, part := range parts {
		// should never happen, consider panic?
		if len(part) == 0 {
			continue
		}
		// is the first character uppercase?
		if part[0] == byte(unicode.ToUpper(rune(part[0]))) {
			indexOfFirstUppercasePart = i
			break
		}
	}
	packageName := strings.Join(parts[:indexOfFirstUppercasePart], ".")
	return packageName, parts[indexOfFirstUppercasePart:]
}

func getFileAndTypeDefForField(project *model.Project, field *protokit.FieldDescriptor) (*descriptor.FileDescriptorProto, *descriptor.DescriptorProto, *descriptor.EnumDescriptorProto, error) {
	packageName, typeNameParts := splitTypeName(field.GetTypeName())
	for _, protoFile := range project.Request.ProtoFile {
		if protoFile.GetPackage() == packageName {
			if len(typeNameParts) == 1 {
				for _, enum := range protoFile.GetEnumType() {
					if enum.GetName() == typeNameParts[0] {
						return protoFile, nil, enum, nil
					}
				}
			}
			for _, msg := range protoFile.GetMessageType() {
				matchMsg, matchEnum, err := searchMessageForNestedType(msg, typeNameParts)
				if err == nil {
					return protoFile, matchMsg, matchEnum, nil
				}
			}
		}
	}

	return nil, nil, nil, errors.Errorf("message %v.%v not found", packageName, typeNameParts)
}

func searchMessageForNestedType(msg *descriptor.DescriptorProto, typeNameParts []string) (*descriptor.DescriptorProto, *descriptor.EnumDescriptorProto, error) {
	switch len(typeNameParts) {
	case 0:
		return nil, nil, errors.Errorf("internal error: ran out of type name parts to try")
	case 1:
		if msg.GetName() == typeNameParts[0] {
			return msg, nil, nil
		}
	default:
		for _, nestedMsg := range msg.GetNestedType() {
			msg, enum, err := searchMessageForNestedType(nestedMsg, typeNameParts[1:])
			if err == nil {
				return msg, enum, nil
			}
		}
		for _, nestedEnum := range msg.GetEnumType() {
			if nestedEnum.GetName() == typeNameParts[1] {
				return nil, nestedEnum, nil
			}
		}
	}
	return nil, nil, errors.Errorf("msg %v does not match type name %v", msg.GetName(), typeNameParts)
}

func (c *templateFunctions) forEachMessage(inFile *protokit.FileDescriptor, messages []*protokit.Descriptor, messageTemplate, enumTemplate string) (string, error) {
	msgTmpl, err := template.New("msgtmpl").Funcs(c.Funcs).Parse(messageTemplate)
	if err != nil {
		return "", err
	}
	enumTmpl, err := template.New("enumtpml").Funcs(c.Funcs).Parse(enumTemplate)
	if err != nil {
		return "", err
	}
	str := ""
	for _, msg := range messages {
		// todo: add parameter to disable this
		// ilackarms: the purpose of this block is to skip
		// messages in the descriptor that are used by proto to represent map types
		if strings.HasSuffix(msg.GetName(), "Entry") &&
			len(msg.GetField()) == 2 &&
			msg.GetField()[0].GetName() == "key" &&
			msg.GetField()[1].GetName() == "value" {
			continue
		}
		renderedMsgString := &bytes.Buffer{}
		if err := msgTmpl.Execute(renderedMsgString, msg); err != nil {
			return "", err
		}
		str += renderedMsgString.String() + "\n"
		if len(msg.GetMessages()) > 0 {
			nested, err := c.forEachMessage(inFile, msg.GetMessages(), messageTemplate, enumTemplate)
			if err != nil {
				return "", err
			}
			str += nested
		}
		// TODO: ilackarms: this might get weird for templates that rely on specifiy enum or msg data
		// for now it works because we only need the name of the type
		for _, enum := range msg.GetEnums() {
			renderedEnumString := &bytes.Buffer{}
			if err := enumTmpl.Execute(renderedEnumString, enum); err != nil {
				return "", err
			}
			str += renderedEnumString.String() + "\n"
		}
	}
	//str = strings.TrimSuffix(str, "\n")
	return str, nil
}
