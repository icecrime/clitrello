package main

import (
	"log"

	"github.com/gbin/goncurses"
)

type Application interface {
	Config() *Config
	Dispatch(Action)
	SwitchState(Screen)
	WindowSize() (height, width int)
}

type EventHandler interface {
	HandleKey(goncurses.Key)
	HandleHTTPResponse(interface{})
}

type Screen interface {
	EventHandler
	Create()
	Destroy()
}

/**
 ******************************************************************************
 */

type Context struct {
	activeScreen Screen
	config       *Config
	mainWindow   *goncurses.Window
	actChannel   chan<- func()
	kbdChannel   <-chan goncurses.Key
}

// Retrieve the application configuration object.
func (context *Context) Config() *Config {
	return context.config
}

// Dispatches a function to be executed in ncurses goroutine.
func (context *Context) Dispatch(action Action) {
	go func(context *Context, screen Screen) {
		response := Execute(context.config, action)
		context.actChannel <- func() {
			if screen == context.activeScreen {
				screen.HandleHTTPResponse(response)
			}
		}
	}(context, context.activeScreen)
}

func (context *Context) SwitchState(newState Screen) {
	if context.activeScreen != nil {
		context.activeScreen.Destroy()
	}

	context.mainWindow.Erase()
	context.mainWindow.Keypad(true)
	context.mainWindow.Refresh()

	context.activeScreen = newState
	context.activeScreen.Create()
}

func (context *Context) WindowSize() (int, int) {
	return context.mainWindow.MaxYX()
}

/**
 ******************************************************************************
 */

func Visual(config *Config) {
	defer func() {
		// Because the recover func is the first function defered, we have the
		// guarantee that proper ncurses deinit will take place in any case of
		// panic in drawing functions.
		if r := recover(); r != nil {
			log.Fatal(r)
		}
	}()

	// Initialize ncurses.
	stdscr, err := goncurses.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer goncurses.End()

	goncurses.Cursor(0)
	goncurses.Echo(false)

	eventLoop(config, stdscr)
}

/**
 ******************************************************************************
 */

func eventLoop(config *Config, window *goncurses.Window) {
	actChannel := make(chan func(), 100)
	kbdChannel := make(chan goncurses.Key, 100)
	winContext := &Context{nil, config, window, actChannel, kbdChannel}

	// Start the keyboard input reading proc.
	ready := make(chan bool)
	input := make(chan goncurses.Key)
	go func(w *goncurses.Window, ch chan<- goncurses.Key) {
		for {
			<-ready
			ch <- w.GetChar()
		}
	}(window, input)

	// Create the initial Screen and start its update loop.
	winContext.SwitchState(NewBoardListScreen(winContext))

	run := true
	for run {
		select {
		case key := <-input:
			switch key {
			case 'q':
				run = false
			default: // forward the key to the active screen handler
				winContext.activeScreen.HandleKey(key)
			}
		case fn := <-actChannel:
			fn()
		case ready <- true:
		}
	}

	close(actChannel)
	close(kbdChannel)
}

/**
 ******************************************************************************
 */

type MenuData struct {
	name string
	desc string
}

func createMenu(items []*goncurses.MenuItem) (menu *goncurses.Menu) {
	var err error
	if menu, err = goncurses.NewMenu(items); err != nil {
		panic(err)
	}
	return menu
}

func createMenuItem(name, desc string) (menuItem *goncurses.MenuItem) {
	var err error
	if menuItem, err = goncurses.NewItem(name, desc); err != nil {
		panic(err)
	}
	return menuItem
}

func createMenuItems(items ...MenuData) []*goncurses.MenuItem {
	menuItems := make([]*goncurses.MenuItem, len(items))
	for i, menuData := range items {
		menuItems[i] = createMenuItem(menuData.name, menuData.desc)
	}
	return menuItems
}

func createPad(h, w int) (pad *goncurses.Pad) {
	var err error
	if pad, err = goncurses.NewPad(h, w); err != nil {
		panic(err)
	}
	return pad
}

func createWindow(h, w, y, x int) (window *goncurses.Window) {
	var err error
	if window, err = goncurses.NewWindow(h, w, y, x); err != nil {
		panic(err)
	}
	return window
}
