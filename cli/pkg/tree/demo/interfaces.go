package demo

type CommandLineRunner interface {
	Run(cmd string, args ...string) error
}
