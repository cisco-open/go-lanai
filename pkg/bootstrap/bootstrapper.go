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

var (
	once                 sync.Once
	bootstrapperInstance *Bootstrapper
)

type ContextOption func(ctx context.Context) context.Context

/**************************
	Singleton Pattern
 **************************/

func bootstrapper() *Bootstrapper {
	once.Do(func() {
		bootstrapperInstance = NewBootstrapper()
	})
	return bootstrapperInstance
}

func Register(m *Module) {
	bootstrapper().Register(m)
}

func AddOptions(options ...fx.Option) {
	bootstrapper().AddOptions(options...)
}

func AddInitialAppContextOptions(options ...ContextOption) {
	bootstrapper().AddInitialAppContextOptions(options...)
}

func AddStartContextOptions(options ...ContextOption) {
	bootstrapper().AddStartContextOptions(options...)
}

func AddStopContextOptions(options ...ContextOption) {
	bootstrapper().AddStopContextOptions(options...)
}

/**************************
	Bootstrapper
 **************************/

// Bootstrapper stores application configurations for bootstrapping
type Bootstrapper struct {
	modules      utils.Set
	adhocModule  *Module
	initCtxOpts  []ContextOption
	startCtxOpts []ContextOption
	stopCtxOpts  []ContextOption
}

// NewBootstrapper create a new Bootstrapper.
// Note: "bootstrap" package uses Singleton patterns for application bootstrap. Calling this function directly is not recommended
// 		 This function is exported for test packages to use
func NewBootstrapper() *Bootstrapper {
	return &Bootstrapper{
		modules:      utils.NewSet(),
		adhocModule:  newAnonymousModule(),
	}
}

func (b *Bootstrapper) Register(m *Module) {
	b.modules.Add(m)
}

func (b *Bootstrapper) AddOptions(options ...fx.Option) {
	b.adhocModule.PriorityOptions = append(b.adhocModule.PriorityOptions, options...)
}

func (b *Bootstrapper) AddInitialAppContextOptions(options ...ContextOption) {
	b.initCtxOpts = append(b.initCtxOpts, options...)
}

func (b *Bootstrapper) AddStartContextOptions(options ...ContextOption) {
	b.startCtxOpts = append(b.startCtxOpts, options...)
}

func (b *Bootstrapper) AddStopContextOptions(options ...ContextOption) {
	b.stopCtxOpts = append(b.stopCtxOpts, options...)
}

// EnableCliRunnerMode implements CliRunnerEnabler
func (b *Bootstrapper) EnableCliRunnerMode(runnerProviders ...interface{}) {
	enableCliRunnerMode(b, runnerProviders)
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
	initModule := InitModule(cliCtx, app)
	miscModules := MiscModules()

	// Decide ad-hoc fx options
	mainModule := newApplicationMainModule()
	for _, o := range priorityOptions {
		mainModule.PriorityOptions = append(mainModule.PriorityOptions, o)
	}

	for _, o := range regularOptions {
		mainModule.Options = append(mainModule.Options, o)
	}

	// Decide modules' fx options
	modules := append(b.modules.Values(), initModule, mainModule, b.adhocModule)
	for _, misc := range miscModules {
		modules = append(modules, misc)
	}
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
	//  1. (Solved)	Support Timeout in bootstrap.Context
	//  2. (Solved) Restore logging
	var cancel context.CancelFunc
	done := app.Done()
	startCtx := app.ctx.Context
	for _, opt := range app.startCtxOpts {
		startCtx = opt(startCtx)
	}

	// This is so that we know that the context in the life cycle hook is the child of bootstrap context
	startCtx, cancel = context.WithTimeout(startCtx, app.StartTimeout())
	defer cancel()

	// log error and exit
	if err := app.Start(startCtx); err != nil {
		logger.WithContext(startCtx).Errorf("Failed to start up: %v", err)
		exit(1)
	}

	// this line blocks until application shutting down
	printSignal(<-done)

	// shutdown sequence
	stopCtx := context.WithValue(app.ctx.Context, ctxKeyStopTime, time.Now().UTC())
	for _, opt := range app.stopCtxOpts {
		stopCtx = opt(stopCtx)
	}

	stopCtx, cancel = context.WithTimeout(stopCtx, app.StopTimeout())
	defer cancel()

	if err := app.Stop(stopCtx); err != nil {
		logger.WithContext(stopCtx).Errorf("Shutdown with Error: %v", err)
		exit(1)
	}
}

func printSignal(signal os.Signal) {
	logger.Infof(strings.ToUpper(signal.String()))
}

func exit(code int) {
	os.Exit(code)
}
