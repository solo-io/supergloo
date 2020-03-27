package table_printing_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTablePrinting(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TablePrinting Suite")
}
