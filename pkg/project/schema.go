package project

type GlobalOptions struct {
	BaseImageRepo    string
	BaseImageVersion string
}
type GoBinaryOutline struct {
	BinaryNameBase string
	ImageName      string
	OutputFile     string
	BinaryDir      string
	// if empty, will write to the default dir with a generated filename
	DockerOutputFilepath string
}
