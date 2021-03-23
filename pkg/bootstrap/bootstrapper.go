package bootstrap

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

var once sync.Once
var bootstrapperInstance *Bootstrapper
var (
	initialContextOptions = []ContextOption{}
	startContextOptions   = []ContextOption{}
	stopContextOptions    = []ContextOption{}
)

type ContextOption func(ctx context.Context) context.Context

type Bootstrapper struct {
	modules utils.Set
}

type App struct {
	*fx.App
	startCtxOptions []ContextOption
	stopCtxOptions  []ContextOption
}

// singleton pattern
func bootstrapper() *Bootstrapper {
	once.Do(func() {
		bootstrapperInstance = &Bootstrapper{
			modules: utils.NewSet(applicationMainModule(), anonymousModule()),
		}
	})
	return bootstrapperInstance
}

func Register(m *Module) {
	b := bootstrapper()
	b.modules.Add(m)
}

func AddOptions(options...fx.Option) {
	m := anonymousModule()
	m.PriorityOptions = append(m.PriorityOptions, options...)
}

func AddInitialAppContextOptions(options...ContextOption) {
	initialContextOptions = append(initialContextOptions, options...)
}

func AddStartContextOptions(options...ContextOption) {
	startContextOptions = append(startContextOptions, options...)
}

func AddStopContextOptions(options...ContextOption) {
	stopContextOptions = append(stopContextOptions, options...)
}

func newApp(cmd *cobra.Command, priorityOptions []fx.Option, regularOptions []fx.Option) *App {
	DefaultModule.PriorityOptions = append(DefaultModule.PriorityOptions, fx.Supply(cmd))
	for _,o := range priorityOptions {
		applicationMainModule().PriorityOptions = append(applicationMainModule().PriorityOptions, o)
	}

	for _,o := range regularOptions {
		applicationMainModule().Options = append(applicationMainModule().Options, o)
	}

	b := bootstrapper()
	modules := b.modules.Values()
	sort.SliceStable(modules, func(i, j int) bool { return modules[i].(*Module).Precedence < modules[j].(*Module).Precedence })

	// add priority options first
	var options []fx.Option
	for _,m := range modules {
		options = append(options, m.(*Module).PriorityOptions...)
	}

	// add other options later
	for _,m := range modules {
		options = append(options, m.(*Module).Options...)
	}

	// update application context before creating the app
	ctx := applicationContext.Context
	for _, opt := range initialContextOptions {
		ctx = opt(ctx)
	}
	applicationContext = applicationContext.withContext(ctx)

	// create App, which will kick off all fx options
	return &App{
		App: fx.New(options...),
		startCtxOptions: startContextOptions,
		stopCtxOptions: stopContextOptions,
	}
}

func (app *App) Run() {
	// to be revised:
	//  1. (Solved)	Support Timeout in bootstrap.Context and make cancellable context as startParent (swap startParent and child)
	//  2. (Solved) Restore logging
	start := time.Now()
	done := app.Done()
	rootCtx := applicationContext.Context
	startParent, cancel := context.WithTimeout(rootCtx, app.StartTimeout())
	for _, opt := range app.startCtxOptions {
		startParent = opt(startParent)
	}
	// This is so that we know that the context in the life cycle hook is the bootstrap context
	startCtx := applicationContext.withContext(startParent)
	defer cancel()

	// log error and exit
	if err := app.Start(startCtx); err != nil {
		logger.WithContext(startCtx).Errorf("Failed to start up: %v", err)
		exit(1)
	}

	// log startup time
	elapsed := time.Now().Sub(start).Truncate(time.Millisecond)
	logger.WithContext(rootCtx).Infof("Started %s after %v", applicationContext.Name(), elapsed)

	// this line blocks until application shutting down
	printSignal(<-done)

	// shutdown sequence
	start = time.Now()
	stopParent, cancel := context.WithTimeout(rootCtx, app.StopTimeout())
	for _, opt := range app.stopCtxOptions {
		stopParent = opt(stopParent)
	}
	stopCtx := applicationContext.withContext(stopParent)
	defer cancel()

	if err := app.Stop(stopCtx); err != nil {
		logger.WithContext(stopCtx).Errorf("Shutdown with Error: %v", err)
		exit(1)
	}

	// log startup time
	elapsed = time.Now().Sub(start).Truncate(time.Millisecond)
	logger.WithContext(rootCtx).Infof("Stopped %s in %v", applicationContext.Name(), elapsed)
}


func printSignal(signal os.Signal) {
	logger.Infof(strings.ToUpper(signal.String()))
}

func exit(code int) {
	os.Exit(code)
}
