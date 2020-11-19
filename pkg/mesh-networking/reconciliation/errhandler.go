package reconciliation

import (
	"reflect"

	"github.com/solo-io/skv2/contrib/pkg/output"
	"github.com/solo-io/skv2/pkg/ezkube"
)

var _ output.ErrorHandler = errHandler{}

type errHandler struct{}

func (e errHandler) HandleWriteError(resource ezkube.Object, err error) {
	val := reflect.ValueOf(resource)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	errs := val.FieldByName("Status").FieldByName("Errors")
	errs.Set(reflect.Append(errs, reflect.ValueOf(err.Error())))
}

func (e errHandler) HandleDeleteError(resource ezkube.Object, err error) {
	val := reflect.ValueOf(resource)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	errs := val.FieldByName("Status").FieldByName("Errors")
	errs.Set(reflect.Append(errs, reflect.ValueOf(err.Error())))
}

func (e errHandler) HandleListError(resource ezkube.Object, err error) {
}
