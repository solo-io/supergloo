package generator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/solo-io/mesh-projects/pkg/project"
)

// make files
// go binary
// docker lists

type GoBinaryOutlineTemplate struct {
	GoBinaryOutline *project.GoBinaryOutline
	Global          *project.GlobalOptions
}
type ServiceManifestOutline struct {
	AppGroup string
	AppName  string
	// if empty, will write to the default dir
	OutputDir string
}
type OutFile struct {
	Filename string
	Content  string
	// if empty, will write to the default dir with a generated file name
	Filepath string
}
type GenBuildOptions struct {
	GlobalOptions           *project.GlobalOptions
	OutputDir               string
	GoBinaries              []*project.GoBinaryOutline
	ServiceManifestOutlines []*ServiceManifestOutline
}

func generateFileContent(genOpts *GenBuildOptions) ([]*OutFile, error) {
	var files []*OutFile
	for _, goBinary := range genOpts.GoBinaries {
		makefileContent, err := generateGoBinaryMakefileContent(goBinary, GoBinaryMakefileBuildTemplate)
		if err != nil {
			return nil, err
		}
		files = append(files, makefileContent)
		dockerfileContent, err := generateGoBinaryDockerfile(&GoBinaryOutlineTemplate{
			GoBinaryOutline: goBinary,
			Global:          genOpts.GlobalOptions,
		}, GoBinaryDockerfileWithCommonBaseImageTemplate)
		if err != nil {
			return nil, err
		}
		files = append(files, dockerfileContent)
	}
	for _, manifestOutline := range genOpts.ServiceManifestOutlines {
		file, err := generateManifestFileContent(manifestOutline, BasicServiceManifestTemplate)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func generateGoBinaryMakefileContent(cfg *project.GoBinaryOutline, tmpl *template.Template) (*OutFile, error) {
	genContent, err := renderTemplate(cfg, tmpl)
	if err != nil {
		return nil, err
	}
	return &OutFile{
		Filename: fmt.Sprintf("%v.%v.partial.makefile",
			tmpl.Name(),
			cfg.BinaryNameBase),
		Content: genContent,
	}, nil
}

func generateGoBinaryDockerfile(cfg *GoBinaryOutlineTemplate, tmpl *template.Template) (*OutFile, error) {
	genContent, err := renderTemplate(cfg, tmpl)
	if err != nil {
		return nil, err
	}
	return &OutFile{
		Filename: fmt.Sprintf("%v.%v.dockerfile",
			tmpl.Name(),
			cfg.GoBinaryOutline.BinaryNameBase,
		),
		Content:  genContent,
		Filepath: cfg.GoBinaryOutline.DockerOutputFilepath,
	}, nil
}

func generateManifestFileContent(cfg *ServiceManifestOutline, tmpl *template.Template) (*OutFile, error) {
	genContent, err := renderTemplate(cfg, tmpl)
	if err != nil {
		return nil, err
	}
	return &OutFile{
		Filename: fmt.Sprintf("%v-%v.yaml",
			tmpl.Name(),
			cfg.AppName,
		),
		Content:  genContent,
		Filepath: cfg.OutputDir,
	}, nil

}
func Generate(genOpts *GenBuildOptions) error {
	files, err := generateFileContent(genOpts)
	if err != nil {
		return err
	}
	for _, f := range files {
		outputFilepath := f.Filepath
		if outputFilepath == "" {
			outputFilepath = filepath.Join(genOpts.OutputDir, f.Filename)
		}
		fmt.Println(outputFilepath)
		if err := os.MkdirAll(filepath.Dir(outputFilepath), 0777); err != nil {
			return err
		}
		if err := ioutil.WriteFile(outputFilepath, []byte(f.Content), 0644); err != nil {
			return err
		}
	}
	return nil
}

func renderTemplate(gbo interface{}, tmpl *template.Template) (string, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, gbo); err != nil {
		return "", err
	}
	return buf.String(), nil
}
