package main

import (
	"code.google.com/p/goncurses"
)

type BoardScreen struct {
	done         bool
	boardId      string
	boardName    string
	httpResponse <-chan []interface{}
	listMenus    []*goncurses.Menu
	listWindows  []*goncurses.Window
	titleWindow  *goncurses.Window
}

func NewBoardScreen(boardId, boardName string) *BoardScreen {
	return &BoardScreen{boardId: boardId, boardName: boardName}
}

func (screen *BoardScreen) Create(context *Context) {
	httpResponse := make(chan []interface{}, 1)
	screen.httpResponse = httpResponse

	go func(resultChannel chan []interface{}, boardId string) {
		resultChannel <- Execute(context.config, GetBoardCardsAction(boardId))
	}(httpResponse, screen.boardId)

	_, x := context.mainWindow.MaxYX()
	screen.listWindows = make([]*goncurses.Window, 0, 10)
	screen.titleWindow = createWindow(3, x, 0, 0)
	screen.titleWindow.Box(0, 0)
	screen.titleWindow.MovePrint(1, 1, " "+screen.boardName+" ")
	screen.titleWindow.Refresh()
}

func (screen *BoardScreen) Update(context *Context) {
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

func (screen *BoardScreen) Destroy() {
	screen.done = true
}

func (screen *BoardScreen) handleKey(context *Context, key goncurses.Key) {
	switch key {
	case '<':
		screen.done = true
		context.actChannel <- func() {
			context.SwitchState(NewBoardListScreen())
		}
	}
}

func (screen *BoardScreen) handleHttpResponse(context *Context, response []interface{}) {
	for i, boardList := range response {
		window := createWindow(12, 40, 4, i*42)
		window.Box(0, 0)
		window.Refresh()
		screen.listWindows = append(screen.listWindows, window)

		listData := boardList.(map[string]interface{})
		listTitle := " " + listData["name"].(string) + " "
		context.mainWindow.MovePrint(4, i*42+2, listTitle)
		context.mainWindow.Refresh()

		menuData := make([]MenuData, 0, len(response))
		if listData["cards"] != nil {
			for _, cardItem := range listData["cards"].([]interface{}) {
				cardData := cardItem.(map[string]interface{})
				menuData = append(menuData, MenuData{cardData["name"].(string), ""})
			}

			sub := window.Derived(10, 38, 1, 1)
			menu := createMenu(createMenuItems(menuData...))
			menu.Format(10, 1)
			menu.Option(goncurses.O_SHOWDESC, false)
			menu.SetWindow(window)
			menu.SubWindow(sub)
			menu.Post()
			screen.listMenus = append(screen.listMenus, menu)

			window.Refresh()
		}
	}
}
