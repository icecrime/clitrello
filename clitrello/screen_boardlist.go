package main

import (
	"github.com/gbin/goncurses"
)

const (
	BOARD_LIST_WIDTH = 50
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

	screen.boardsWdw = createWindow(2, BOARD_LIST_WIDTH, 0, 0)
	screen.boardsWdw.Box(0, 0)
	screen.boardsWdw.MovePrint(0, 2, " My Boards ")
	screen.boardsWdw.Refresh()
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

func (screen *BoardListScreen) HandleHTTPResponse(response interface{}) {
	boardList := response.([]*BoardInfo)
	menuData := make([]MenuData, len(boardList))
	for i, userBoard := range boardList {
		menuData[i] = MenuData{userBoard.Name, userBoard.Id}
	}

	windowHeight := len(menuData) + 2
	if height, _ := screen.application.WindowSize(); windowHeight > height {
		windowHeight = height
	}

	screen.boardsWdw.Clear()
	screen.boardsWdw.Resize(windowHeight, BOARD_LIST_WIDTH)
	screen.boardsWdw.Box(0, 0)
	screen.boardsWdw.MovePrint(0, 2, " My Boards ")

	screen.boardsMenuWdw = screen.boardsWdw.Derived(windowHeight-2, BOARD_LIST_WIDTH-2, 1, 1)
	screen.boardsMenu = createMenu(createMenuItems(menuData...))
	screen.boardsMenu.Format(windowHeight-2, 1)
	screen.boardsMenu.Option(goncurses.O_SHOWDESC, false)
	screen.boardsMenu.SetWindow(screen.boardsWdw)
	screen.boardsMenu.SubWindow(screen.boardsMenuWdw)
	screen.boardsMenu.Post()

	screen.boardsWdw.Refresh()
	screen.boardsMenuWdw.Refresh()

	screen.ready = true
}
