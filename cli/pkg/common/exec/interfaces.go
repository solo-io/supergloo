package exec

//go:generate mockgen -destination ./mocks/interfaces.go -source ./interfaces.go

type Runner interface {
	Run(cmd string, args ...string) error
}
