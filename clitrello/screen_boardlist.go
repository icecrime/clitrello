package main

import (
	"code.google.com/p/goncurses"
)

type BoardListScreen struct {
	done            bool
	menuReady       bool
	httpResponse    <-chan []interface{}
	boardsWdw       *goncurses.Window
	boardsMenuWdw   *goncurses.Window
	boardsMenu      *goncurses.Menu
	boardsMenuItems []*goncurses.MenuItem
}

func NewBoardListScreen() *BoardListScreen {
	return &BoardListScreen{}
}

func (screen *BoardListScreen) Create(context *Context) {
	httpResponse := make(chan []interface{}, 1)
	screen.httpResponse = httpResponse

	go func(resultChannel chan []interface{}) {
		resultChannel <- Execute(context.config, ListBoardsAction())
	}(httpResponse)

	screen.boardsWdw = createWindow(12, 25, 0, 0)
	screen.boardsWdw.Box(0, 0)
	screen.boardsWdw.MovePrint(0, 1, " My Boards ")
	screen.boardsWdw.Refresh()
	screen.boardsMenuWdw = screen.boardsWdw.Derived(10, 23, 1, 1)
	screen.boardsMenuWdw.Keypad(true)
}

func (screen *BoardListScreen) Update(context *Context) {
	var key goncurses.Key
	for kbdChannelActive := true; kbdChannelActive && !screen.done; {
		select {
		case key, kbdChannelActive = <-context.kbdChannel:
			screen.handleKey(context, key)
		case response := <-screen.httpResponse:
			screen.handleHttpResponse(context, response)
		}
	}
}

func (screen *BoardListScreen) Destroy() {
	screen.done = true
	screen.boardsMenu.UnPost()
	for _, menuItem := range screen.boardsMenuItems {
		menuItem.Free()
	}
	screen.boardsMenu.Free()
}

func (screen *BoardListScreen) handleKey(context *Context, key goncurses.Key) {
	if screen.menuReady {
		switch key {
		case goncurses.KEY_DOWN:
			context.actChannel <- func() {
				screen.boardsMenu.Driver(goncurses.REQ_DOWN)
				screen.boardsMenuWdw.Refresh()
			}
		case goncurses.KEY_UP:
			context.actChannel <- func() {
				screen.boardsMenu.Driver(goncurses.REQ_UP)
				screen.boardsMenuWdw.Refresh()
			}
		case goncurses.KEY_RETURN:
			screen.done = true
			context.actChannel <- func() {
				active := screen.boardsMenu.Current(nil)
				context.SwitchState(NewBoardScreen(active.Description(), active.Name()))
			}
		}
	}
}

func (screen *BoardListScreen) handleHttpResponse(context *Context, response []interface{}) {
	context.actChannel <- func() {
		menuData := make([]MenuData, 0, len(response))
		for _, userBoard := range response {
			boardData := userBoard.(map[string]interface{})
			menuData = append(menuData, MenuData{boardData["name"].(string), boardData["id"].(string)})
		}

		screen.boardsMenuItems = createMenuItems(menuData...)
		screen.boardsMenu = createMenu(screen.boardsMenuItems)
		screen.boardsMenu.Format(10, 1)
		screen.boardsMenu.Option(goncurses.O_SHOWDESC, false)
		screen.boardsMenu.SetWindow(screen.boardsWdw)
		screen.boardsMenu.SubWindow(screen.boardsMenuWdw)
		screen.boardsMenu.Post()
		screen.boardsMenuWdw.Refresh()

		screen.menuReady = true
	}
}
