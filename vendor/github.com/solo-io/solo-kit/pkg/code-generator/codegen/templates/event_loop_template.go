package templates

import (
	"text/template"
)

var ResourceGroupEventLoopTemplate = template.Must(template.New("resource_group_event_loop").Funcs(Funcs).Parse(`package {{ .Project.ProjectConfig.Version }}

import (
	"context"

	"go.opencensus.io/trace"

	"github.com/hashicorp/go-multierror"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/utils/errutils"
)

type {{ .GoName }}Syncer interface {
	Sync(context.Context, *{{ .GoName }}Snapshot) error
}

type {{ .GoName }}Syncers []{{ .GoName }}Syncer

func (s {{ .GoName }}Syncers) Sync(ctx context.Context, snapshot *{{ .GoName }}Snapshot) error {
	var multiErr *multierror.Error
	for _, syncer := range s {
		if err := syncer.Sync(ctx, snapshot); err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return multiErr.ErrorOrNil()
}

type {{ .GoName }}EventLoop interface {
	Run(namespaces []string, opts clients.WatchOpts) (<-chan error, error)
}

type {{ lower_camel .GoName }}EventLoop struct {
	emitter {{ .GoName }}Emitter
	syncer  {{ .GoName }}Syncer
}

func New{{ .GoName }}EventLoop(emitter {{ .GoName }}Emitter, syncer {{ .GoName }}Syncer) {{ .GoName }}EventLoop {
	return &{{ lower_camel .GoName }}EventLoop{
		emitter: emitter,
		syncer:  syncer,
	}
}

func (el *{{ lower_camel .GoName }}EventLoop) Run(namespaces []string, opts clients.WatchOpts) (<-chan error, error) {
	opts = opts.WithDefaults()
	opts.Ctx = contextutils.WithLogger(opts.Ctx, "{{ .Project.ProjectConfig.Version }}.event_loop")
	logger := contextutils.LoggerFrom(opts.Ctx)
	logger.Infof("event loop started")

	errs := make(chan error)

	watch, emitterErrs, err := el.emitter.Snapshots(namespaces, opts)
	if err != nil {
		return nil, errors.Wrapf(err, "starting snapshot watch")
	}
	go errutils.AggregateErrs(opts.Ctx, errs, emitterErrs, "{{ .Project.ProjectConfig.Version }}.emitter errors")
	go func() {
		// create a new context for each loop, cancel it before each loop
		var cancel context.CancelFunc = func() {}
		// use closure to allow cancel function to be updated as context changes
		defer func() { cancel() }()
		for {
			select {
			case snapshot, ok := <-watch:
				if !ok {
					return
				}
				// cancel any open watches from previous loop
				cancel()

				ctx, span := trace.StartSpan(opts.Ctx, "{{ .Name }}.EventLoopSync")
				ctx, canc := context.WithCancel(ctx)
				cancel = canc
				err := el.syncer.Sync(ctx, snapshot)
				span.End()

				if err != nil {
					select {
					case errs <- err:
					default:
						logger.Errorf("write error channel is full! could not propagate err: %v", err)
					}
				}
			case <-opts.Ctx.Done():
				return
			}
		}
	}()
	return errs, nil
}
`))
