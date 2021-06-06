package bootstrap

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"go.uber.org/fx"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

/**************************
	Bootstrapper
 **************************/

var once sync.Once
var bootstrapperInstance *Bootstrapper

type ContextOption func(ctx context.Context) context.Context

type Bootstrapper struct {
	modules      utils.Set
	adhocModule  *Module
	initCtxOpts  []ContextOption
	startCtxOpts []ContextOption
	stopCtxOpts  []ContextOption
}

// singleton pattern
func bootstrapper() *Bootstrapper {
	once.Do(func() {
		bootstrapperInstance = &Bootstrapper{
			modules:      utils.NewSet(),
			adhocModule:  newAnonymousModule(),
			initCtxOpts:  []ContextOption{},
			startCtxOpts: []ContextOption{},
			stopCtxOpts:  []ContextOption{},
		}
	})
	return bootstrapperInstance
}

func Register(m *Module) {
	b := bootstrapper()
	b.modules.Add(m)
}

func AddOptions(options ...fx.Option) {
	b := bootstrapper()
	b.adhocModule.PriorityOptions = append(b.adhocModule.PriorityOptions, options...)
}

func AddInitialAppContextOptions(options ...ContextOption) {
	b := bootstrapper()
	b.initCtxOpts = append(b.initCtxOpts, options...)
}

func AddStartContextOptions(options ...ContextOption) {
	b := bootstrapper()
	b.startCtxOpts = append(b.startCtxOpts, options...)
}

func AddStopContextOptions(options ...ContextOption) {
	b := bootstrapper()
	b.stopCtxOpts = append(b.stopCtxOpts, options...)
}

func (b *Bootstrapper) NewApp(cliCtx *CliExecContext, priorityOptions []fx.Option, regularOptions []fx.Option) *App {
	// create App
	app := &App{
		ctx:          NewApplicationContext(),
		startCtxOpts: b.startCtxOpts,
		stopCtxOpts:  b.stopCtxOpts,
	}

	// update application context before creating the app
	ctx := app.ctx.Context
	for _, opt := range b.initCtxOpts {
		ctx = opt(ctx)
	}
	app.ctx = app.ctx.withContext(ctx)

	// Decide default module
	defaultModule := DefaultModule(cliCtx, app)

	// Decide ad-hoc fx options
	mainModule := newApplicationMainModule()
	for _, o := range priorityOptions {
		mainModule.PriorityOptions = append(mainModule.PriorityOptions, o)
	}

	for _, o := range regularOptions {
		mainModule.Options = append(mainModule.Options, o)
	}

	// Decide modules' fx options
	modules := append(b.modules.Values(), defaultModule, mainModule, b.adhocModule)
	sort.SliceStable(modules, func(i, j int) bool { return modules[i].(*Module).Precedence < modules[j].(*Module).Precedence })

	// add priority options first
	var options []fx.Option
	for _, m := range modules {
		options = append(options, m.(*Module).PriorityOptions...)
	}

	// add other options later
	for _, m := range modules {
		options = append(options, m.(*Module).Options...)
	}

	// create fx.App, which will kick off all fx options
	app.App = fx.New(options...)
	return app
}

/**************************
	Application
 **************************/

type App struct {
	*fx.App
	ctx          *ApplicationContext
	startCtxOpts []ContextOption
	stopCtxOpts  []ContextOption
}

// EagerGetApplicationContext returns the global ApplicationContext before it becomes available for dependency injection
// Important: packages should typically get ApplicationContext via fx's dependency injection,
//			  which internal application config are guaranteed.
//			  Only packages involved in priority bootstrap (appconfig, consul, vault, etc)
//			  should use this function for logging purpose
func (app *App) EagerGetApplicationContext() *ApplicationContext {
	return app.ctx
}

func (app *App) Run() {
	// to be revised:
	//  1. (Solved)	Support Timeout in bootstrap.Context and make cancellable context as startParent (swap startParent and child)
	//  2. (Solved) Restore logging
	start := time.Now()
	done := app.Done()
	rootCtx := app.ctx.Context
	startParent, cancel := context.WithTimeout(rootCtx, app.StartTimeout())
	for _, opt := range app.startCtxOpts {
		startParent = opt(startParent)
	}
	// This is so that we know that the context in the life cycle hook is the bootstrap context
	startCtx := app.ctx.withContext(startParent)
	defer cancel()

	// log error and exit
	if err := app.Start(startCtx); err != nil {
		logger.WithContext(startCtx).Errorf("Failed to start up: %v", err)
		exit(1)
	}

	// log startup time
	elapsed := time.Now().Sub(start).Truncate(time.Millisecond)
	logger.WithContext(rootCtx).Infof("Started %s after %v", app.ctx.Name(), elapsed)

	// this line blocks until application shutting down
	printSignal(<-done)

	// shutdown sequence
	start = time.Now()
	stopParent, cancel := context.WithTimeout(rootCtx, app.StopTimeout())
	for _, opt := range app.stopCtxOpts {
		stopParent = opt(stopParent)
	}
	stopCtx := app.ctx.withContext(stopParent)
	defer cancel()

	if err := app.Stop(stopCtx); err != nil {
		logger.WithContext(stopCtx).Errorf("Shutdown with Error: %v", err)
		exit(1)
	}

	// log startup time
	elapsed = time.Now().Sub(start).Truncate(time.Millisecond)
	logger.WithContext(rootCtx).Infof("Stopped %s in %v", app.ctx.Name(), elapsed)
}

func printSignal(signal os.Signal) {
	logger.Infof(strings.ToUpper(signal.String()))
}

func exit(code int) {
	os.Exit(code)
}
