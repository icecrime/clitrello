package main

import (
	"github.com/gbin/goncurses"
)

const (
	CARD_LIST_SPACING = 2
	CARD_LIST_WIDTH   = 50
	TOTAL_LIST_WIDTH  = CARD_LIST_WIDTH + CARD_LIST_SPACING

	TITLE_WINDOW_HEIGHT = 3
)

type BoardScreen struct {
	application Application
	done        bool
	ready       bool

	boardId     string
	boardName   string
	listsPad    *goncurses.Pad
	titleWindow *goncurses.Window

	active     int
	firstDrawn int
	cards      []*CardList
}

func NewBoardScreen(application Application, boardId, boardName string) *BoardScreen {
	return &BoardScreen{application: application, boardId: boardId, boardName: boardName}
}

func (screen *BoardScreen) Create() {
	// Start an asynchronous operation to retrieve the board content from the
	// Trello API.
	screen.application.Dispatch(NewGetBoardCardsAction(screen.boardId))

	// Create the title window: we wait for the HTTP response to create the
	// list windows.
	_, width := screen.application.WindowSize()
	screen.titleWindow = createWindow(TITLE_WINDOW_HEIGHT, width, 0, 0)
	screen.titleWindow.Box(0, 0)
	screen.titleWindow.MovePrint(1, 2, screen.boardName)
	screen.titleWindow.Refresh()
}

func (screen *BoardScreen) Destroy() {
	screen.done = true
	for _, cardList := range screen.cards {
		cardList.Destroy()
	}

	if screen.titleWindow != nil {
		screen.titleWindow.Delete()
	}

	if screen.listsPad != nil {
		screen.listsPad.Delete()
	}
}

func (screen *BoardScreen) HandleKey(key goncurses.Key) {
	if screen.done || !screen.ready {
		return
	}

	switch key {
	case goncurses.KEY_LEFT:
		screen.navigateBetweenLists(screen.preceedingIndex)
	case goncurses.KEY_RIGHT:
		screen.navigateBetweenLists(screen.succeedingIndex)
	case goncurses.KEY_DOWN:
		screen.cards[screen.active].Driver(goncurses.REQ_DOWN)
	case goncurses.KEY_UP:
		screen.cards[screen.active].Driver(goncurses.REQ_UP)
	case goncurses.KEY_RETURN:
		screen.cards[screen.active].ViewCard()
	case '<':
		screen.done = true
		screen.application.SwitchState(NewBoardListScreen(screen.application))
		return
	}

	screen.renderPad()
}

func (screen *BoardScreen) HandleHTTPResponse(response interface{}) {
	boardLists := response.([]*TrelloList)

	// Create a pad which will host the different card list windows.
	height, _ := screen.application.WindowSize()
	screen.listsPad = createPad(height-TITLE_WINDOW_HEIGHT, len(boardLists)*TOTAL_LIST_WIDTH-CARD_LIST_SPACING)

	screen.cards = make([]*CardList, len(boardLists))
	for i, item := range boardLists {
		screen.cards[i] = NewCardList(screen, item, i*TOTAL_LIST_WIDTH)
	}

	// Give focus to the first list
	screen.setActiveList(0)

	// Render and Mark the screen as ready
	screen.renderPad()
	screen.ready = true
}

func (screen *BoardScreen) horizontalScroll() {
	_, width := screen.application.WindowSize()
	_, x := screen.cards[screen.active].window.YX()

	// Scroll to the left until the active list left corner is visible.
	// Scroll to the right until the active list right corner is visible.
	firstDrawn := screen.firstDrawn
	for firstDrawn > 0 && x < firstDrawn*TOTAL_LIST_WIDTH {
		firstDrawn--
	}
	for firstDrawn < len(screen.cards) && x > width+(firstDrawn-1)*TOTAL_LIST_WIDTH {
		firstDrawn++
	}

	// Avoid unecessary redraw if scrolling wasn't necessary. Note that if we
	// dont clear the pad _before_ rendering a new portion, we'll get artifacts
	// on the screen.
	if firstDrawn != screen.firstDrawn {
		screen.clearPad()
		screen.firstDrawn = firstDrawn
		screen.renderPad()
	}
}

func (screen *BoardScreen) navigateBetweenLists(wayFunc func() int) {
	listCount := len(screen.cards)
	if nextIndex := wayFunc(); nextIndex >= 0 && nextIndex < listCount {
		screen.setActiveList(nextIndex)
		screen.horizontalScroll()
	}
}

func (screen *BoardScreen) clearPad() {
	height, width := screen.application.WindowSize()
	screen.listsPad.Clear()
	screen.listsPad.Refresh(0, screen.firstDrawn*(CARD_LIST_WIDTH+2), 4, 0, height-4, width-2)
}

func (screen *BoardScreen) renderPad() {
	height, width := screen.application.WindowSize()
	for _, cardList := range screen.cards {
		cardList.Clear()
		cardList.Draw()
	}
	screen.listsPad.Refresh(0, screen.firstDrawn*(CARD_LIST_WIDTH+2), 4, 0, height-4, width-2)
}

func (screen *BoardScreen) setActiveList(activeIndex int) {
	screen.cards[screen.active].Focus(false)
	screen.active = activeIndex
	screen.cards[screen.active].Focus(true)
}

func (screen *BoardScreen) preceedingIndex() (index int) {
	index = screen.active - 1
	for index >= 0 && screen.cards[index].Count() == 0 {
		index--
	}
	return
}

func (screen *BoardScreen) succeedingIndex() (index int) {
	index = screen.active + 1
	for index < len(screen.cards) && screen.cards[index].Count() == 0 {
		index++
	}
	return
}

/**
 ******************************************************************************
 */

type CardList struct {
	application Application

	focus bool
	title string

	window     *goncurses.Window
	menu       *goncurses.Menu
	menuWindow *goncurses.Window

	list  *TrelloList
	cards map[string]*TrelloCard
}

func NewCardList(screen *BoardScreen, listDetail *TrelloList, x int) *CardList {
	list := &CardList{application: screen.application, list: listDetail}

	// We want to create a curses menu which the list of cards: this is nothing
	// more than a transtyping nightmare.
	nbCards := len(listDetail.Cards)
	menuItems := make([]MenuData, nbCards)
	list.cards = make(map[string]*TrelloCard, nbCards)
	for i, cardItem := range listDetail.Cards {
		list.cards[cardItem.Name] = cardItem
		menuItems[i] = MenuData{cardItem.Name, ""}
	}

	// Create the curses visual elements.
	list.title = " " + listDetail.Name + " "
	list.window = screen.listsPad.Derived(nbCards+2, CARD_LIST_WIDTH, 0, x)
	list.menuWindow = list.window.Derived(nbCards, CARD_LIST_WIDTH-3, 1, 1)

	list.menu = createMenu(createMenuItems(menuItems...))
	list.menu.Format(nbCards, 1)
	list.menu.Mark(" ")
	list.menu.Option(goncurses.O_SHOWDESC, false)
	list.menu.SetWindow(list.window)
	list.menu.SubWindow(list.menuWindow)
	list.menu.Post()

	list.window.Box(0, 0)
	list.window.MovePrint(0, 2, list.title)

	list.Focus(false)
	return list
}

func (list *CardList) Clear() {
	list.menu.UnPost()
	list.window.Clear()
	list.window.Refresh()
}

func (list *CardList) Count() int {
	return len(list.cards)
}

func (list *CardList) Destroy() {
	for _, menuItem := range list.menu.Items() {
		menuItem.Free()
	}
	list.menu.Free()
	list.menuWindow.Delete()
	list.window.Delete()
}

func (list *CardList) Draw() {
	list.menu.UnPost()
	list.window.Box(0, 0)
	list.window.MovePrint(0, 2, list.title)
	list.menu.Post()
}

func (list *CardList) Driver(daction int) {
	list.menu.Driver(daction)
	list.window.Refresh()
}

func (list *CardList) Focus(focus bool) {
	if list.focus = focus; list.focus {
		list.menu.Mark("-")
		list.menu.SetForeground(goncurses.Char(list.menu.Foreground() | goncurses.A_REVERSE))
	} else {
		list.menu.Mark(" ")
		list.menu.SetForeground(goncurses.Char(list.menu.Foreground() & ^goncurses.A_REVERSE))
	}
}

func (_ *CardList) ViewCard() {
}
