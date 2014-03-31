package main

import (
	"code.google.com/p/goncurses"
)

type BoardListScreen struct {
	application Application
	done        bool
	ready       bool

	boardsWdw       *goncurses.Window
	boardsMenuWdw   *goncurses.Window
	boardsMenu      *goncurses.Menu
	boardsMenuItems []*goncurses.MenuItem
}

func NewBoardListScreen(application Application) *BoardListScreen {
	return &BoardListScreen{application: application}
}

func (screen *BoardListScreen) Create() {
	screen.application.Dispatch(NewListBoardsAction())

	screen.boardsWdw = createWindow(12, 25, 0, 0)
	screen.boardsWdw.Box(0, 0)
	screen.boardsWdw.MovePrint(0, 1, " My Boards ")
	screen.boardsWdw.Refresh()
	screen.boardsMenuWdw = screen.boardsWdw.Derived(10, 23, 1, 1)
	screen.boardsMenuWdw.Keypad(true)
}

func (screen *BoardListScreen) Destroy() {
	screen.done = true
	screen.boardsMenu.UnPost()
	for _, menuItem := range screen.boardsMenu.Items() {
		menuItem.Free()
	}
	screen.boardsMenu.Free()
	screen.boardsMenuWdw.Delete()
	screen.boardsWdw.Delete()
}

func (screen *BoardListScreen) HandleKey(key goncurses.Key) {
	if screen.done || !screen.ready {
		return
	}

	switch key {
	case goncurses.KEY_DOWN:
		screen.boardsMenu.Driver(goncurses.REQ_DOWN)
		screen.boardsMenuWdw.Refresh()
	case goncurses.KEY_UP:
		screen.boardsMenu.Driver(goncurses.REQ_UP)
		screen.boardsMenuWdw.Refresh()
	case goncurses.KEY_RETURN:
		screen.done = true
		active := screen.boardsMenu.Current(nil)
		nextScreen := NewBoardScreen(screen.application, active.Description(), active.Name())
		screen.application.SwitchState(nextScreen)
	}
}

func (screen *BoardListScreen) HandleHTTPResponse(response []interface{}) {
	menuData := make([]MenuData, 0, len(response))
	for _, userBoard := range response {
		boardData := userBoard.(map[string]interface{})
		menuData = append(menuData, MenuData{boardData["name"].(string), boardData["id"].(string)})
	}

	boardsMenuItems := createMenuItems(menuData...)
	screen.boardsMenu = createMenu(boardsMenuItems)
	screen.boardsMenu.Format(10, 1)
	screen.boardsMenu.Option(goncurses.O_SHOWDESC, false)
	screen.boardsMenu.SetWindow(screen.boardsWdw)
	screen.boardsMenu.SubWindow(screen.boardsMenuWdw)
	screen.boardsMenu.Post()
	screen.boardsMenuWdw.Refresh()

	screen.ready = true
}
