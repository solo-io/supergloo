package errhandlers

import (
	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
)

// this file contains ErrorHandlers for handling errors created when writing output snapshots

type AppendingErrHandler struct {
	errs error
}

func (a AppendingErrHandler) HandleWriteError(resource ezkube.Object, err error) {
	a.errs = multierror.Append(a.errs, eris.Wrapf(err, "writing resource %v failed", sets.Key(resource)))
}

func (a AppendingErrHandler) HandleDeleteError(resource ezkube.Object, err error) {
	a.errs = multierror.Append(a.errs, eris.Wrapf(err, "deleting resource %v failed", sets.Key(resource)))
}

func (a AppendingErrHandler) HandleListError(err error) {
	a.errs = multierror.Append(a.errs, eris.Wrapf(err, "listing failed"))
}

// returns the errors collected by the handler
func (a AppendingErrHandler) Errors() error {
	return a.errs
}
