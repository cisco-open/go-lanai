package bootstrap

import (
	"context"
	"fmt"
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
			modules: []*Module{anonymousModule()},
		}
	})
	return bootstrapperInstance
}

func Register(m *Module) {
	b := bootstrapper()
	b.modules = append(b.modules, m)
}

func AddProvides(options...fx.Option) {
	m := anonymousModule()
	m.Provides = append(m.Provides, options...)
}

func AddInvokes(options...fx.Option) {
	m := anonymousModule()
	m.Invokes = append(m.Invokes, options...)
}

func NewApp(adhocOptions...fx.Option) *App {
	b := bootstrapper()
	sort.SliceStable(b.modules, func(i, j int) bool { return b.modules[i].Precedence < b.modules[j].Precedence })
	// add Provides first
	var options = adhocOptions
	for _,m := range b.modules {
		options = append(options, m.Provides...)
	}

	// add Invokes at the end
	for _,m := range b.modules {
		options = append(options, m.Invokes...)
	}
	return &App{App: fx.New(options...)}
}

func (app *App) Run() {
	// TODO to be revised:
	//  1. (Solved)	Support Timeout in bootstrap.Context and make cancellable context as startParent (swap startParent and child)
	//  2. Restore logging
	done := app.Done()
	startParent, cancel := context.WithTimeout(context.Background(), app.StartTimeout())
	startCtx := bootstrapContext.UpdateParent(startParent)
	defer cancel()

	if err := app.Start(startCtx); err != nil {
		//app.logger.Fatalf("ERROR\t\tFailed to start: %v", err)
		fmt.Printf("ERROR\t\tFailed to start up: %v\n", err)
		exit(1)
	}

	printSignal(<-done)
	//app.logger.PrintSignal(<-done)

	stopParent, cancel := context.WithTimeout(context.Background(), app.StopTimeout())
	stopCtx := bootstrapContext.UpdateParent(stopParent)
	defer cancel()

	if err := app.Stop(stopCtx); err != nil {
		//app.logger.Fatalf("ERROR\t\tFailed to stop cleanly: %v", err)
		fmt.Printf("ERROR\t\tFailed to gracefully shutdown: %v\n", err)
		exit(1)
	}
}

func printSignal(signal os.Signal) {
	fmt.Println(strings.ToUpper(signal.String()))
}

func exit(code int) {
	os.Exit(code)
}
