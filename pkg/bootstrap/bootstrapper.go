package bootstrap

import (
	"context"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"os"
	"sort"
	"strings"
	"sync"
)

var once sync.Once
var bootstrapperInstance *Bootstrapper

type Bootstrapper struct {
	modules []*Module
}

type App struct {
	*fx.App
}

// singleton pattern
func bootstrapper() *Bootstrapper {
	once.Do(func() {
		bootstrapperInstance = &Bootstrapper{
			modules: []*Module{applicationMainModule(), anonymousModule()},
		}
	})
	return bootstrapperInstance
}

func Register(m *Module) {
	b := bootstrapper()
	b.modules = append(b.modules, m)
}

func AddOptions(options...fx.Option) {
	m := anonymousModule()
	m.PriorityOptions = append(m.PriorityOptions, options...)
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
	sort.SliceStable(b.modules, func(i, j int) bool { return b.modules[i].Precedence < b.modules[j].Precedence })

	// add priority options first
	var options []fx.Option
	for _,m := range b.modules {
		options = append(options, m.PriorityOptions...)
	}

	// add other options later
	for _,m := range b.modules {
		options = append(options, m.Options...)
	}

	return &App{App: fx.New(options...)}
}

func (app *App) Run() {
	// TODO to be revised:
	//  1. (Solved)	Support Timeout in bootstrap.Context and make cancellable context as startParent (swap startParent and child)
	//  2. Restore logging
	done := app.Done()
	startParent, cancel := context.WithTimeout(context.Background(), app.StartTimeout())
	startCtx := applicationContext.updateParent(startParent) //This is so that we know that the context in the life cycle hook is the bootstrap context
	defer cancel()

	if err := app.Start(startCtx); err != nil {
		logger.WithContext(startCtx).Errorf("Failed to start up: %v", err)
		exit(1)
	}

	printSignal(<-done)
	//app.logger.PrintSignal(<-done)

	stopParent, cancel := context.WithTimeout(context.Background(), app.StopTimeout())
	stopCtx := applicationContext.updateParent(stopParent)
	defer cancel()

	if err := app.Stop(stopCtx); err != nil {
		logger.WithContext(stopCtx).Errorf("Failed to gracefully shutdown: %v", err)
		exit(1)
	}
}


func printSignal(signal os.Signal) {
	logger.Infof(strings.ToUpper(signal.String()))
}

func exit(code int) {
	os.Exit(code)
}
