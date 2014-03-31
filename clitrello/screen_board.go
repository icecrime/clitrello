package main

import (
	"code.google.com/p/goncurses"
)

type BoardScreen struct {
	done         bool
	boardId      string
	boardName    string
	httpResponse <-chan []interface{}
	titleWindow  *goncurses.Window

	activeList int
	cardLists  []*CardList
}

func NewBoardScreen(boardId, boardName string) *BoardScreen {
	return &BoardScreen{boardId: boardId, boardName: boardName}
}

func (screen *BoardScreen) Create(context *Context) {
	httpResponse := make(chan []interface{}, 1)
	go func(resultChannel chan []interface{}, boardId string) {
		resultChannel <- Execute(context.config, GetBoardCardsAction(boardId))
	}(httpResponse, screen.boardId)

	screen.httpResponse = httpResponse

	_, x := context.mainWindow.MaxYX()
	screen.titleWindow = createWindow(3, x, 0, 0)
	screen.titleWindow.Box(0, 0)
	screen.titleWindow.MovePrint(1, 1, " "+screen.boardName+" ")
	screen.titleWindow.Refresh()
}

func (screen *BoardScreen) Destroy() {
	screen.done = true
	for _, cardList := range screen.cardLists {
		cardList.Destroy()
	}
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

func (screen *BoardScreen) handleKey(context *Context, key goncurses.Key) {
	switch key {
	case goncurses.KEY_LEFT:
		context.actChannel <- func() {
			screen.switchActiveList(screen.activeList - 1)
		}
	case goncurses.KEY_RIGHT:
		context.actChannel <- func() {
			screen.switchActiveList(screen.activeList + 1)
		}
	case goncurses.KEY_DOWN:
		context.actChannel <- func() {
			screen.cardLists[screen.activeList].Driver(goncurses.REQ_DOWN)
		}
	case goncurses.KEY_UP:
		context.actChannel <- func() {
			screen.cardLists[screen.activeList].Driver(goncurses.REQ_UP)
		}
	case goncurses.KEY_RETURN:
		context.actChannel <- func() {
			screen.cardLists[screen.activeList].ViewCard()
		}
	case '<':
		screen.done = true
		context.actChannel <- func() {
			context.SwitchState(NewBoardListScreen())
		}
	}
}

func (screen *BoardScreen) handleHttpResponse(context *Context, response []interface{}) {
	var startX int
	for _, cardList := range response {
		cardData := cardList.(map[string]interface{})
		cardList := NewCardList(context, cardData, 4, startX)
		screen.cardLists = append(screen.cardLists, cardList)
		startX += LIST_WIDTH + 2
	}

	// Give focus to the first list
	screen.switchActiveList(0)
}

func (screen *BoardScreen) switchActiveList(activeIndex int) {
	if activeIndex >= 0 && activeIndex < len(screen.cardLists) {
		screen.cardLists[screen.activeList].Focus(false)
		screen.activeList = activeIndex
		screen.cardLists[screen.activeList].Focus(true)
	}
}

/**
 ******************************************************************************
 */

const (
	LIST_WIDTH = 50
)

type CardInfo map[string]interface{}

type CardList struct {
	focus      bool
	window     *goncurses.Window
	menu       *goncurses.Menu
	menuWindow *goncurses.Window
	listData   map[string]interface{}
	cardsData  map[string]CardInfo
}

func NewCardList(context *Context, listData map[string]interface{}, y, x int) *CardList {
	// Extracts the cards elements from the returned JSON object.
	cardsInfo := listData["cards"].([]interface{})
	menuItems := make([]MenuData, 0, len(cardsInfo))
	cardsData := make(map[string]CardInfo, len(cardsInfo))
	for _, cardItem := range cardsInfo {
		cardData := cardItem.(map[string]interface{})
		cardName := cardData["name"].(string)
		cardsData[cardName] = cardData
		menuItems = append(menuItems, MenuData{cardName, ""})
	}

	// Create the curses visual elements.
	list := &CardList{listData: listData, cardsData: cardsData}
	list.window = createWindow(len(cardsData)+2, LIST_WIDTH, y, x)
	list.window.Box(0, 0)
	list.menuWindow = list.window.Derived(len(cardsData), LIST_WIDTH-3, 1, 1)

	list.menu = createMenu(createMenuItems(menuItems...))
	list.menu.Format(len(cardsData), 1)
	list.menu.Mark(" ")
	list.menu.Option(goncurses.O_SHOWDESC, false)
	list.menu.SetWindow(list.window)
	list.menu.SubWindow(list.menuWindow)
	list.menu.Post()

	listTitle := " " + listData["name"].(string) + " "
	list.window.MovePrint(0, 2, listTitle)
	list.window.Refresh()

	list.Focus(false)
	return list
}

func (list *CardList) Destroy() {
	list.menu.UnPost()
	for _, menuItem := range list.menu.Items() {
		menuItem.Free()
	}
	list.menu.Free()
	list.menuWindow.Delete()
	list.window.Delete()
}

func (list *CardList) Driver(daction int) {
	list.menu.Driver(daction)
	list.window.Refresh()
}

func (list *CardList) Focus(focus bool) {
	list.focus = focus
	if list.focus {
		list.menu.Mark("*")
		list.menu.SetForeground(goncurses.Char(list.menu.Foreground() | goncurses.A_REVERSE))
	} else {
		list.menu.Mark(" ")
		list.menu.SetForeground(goncurses.Char(list.menu.Foreground() & ^goncurses.A_REVERSE))
	}
	list.window.Refresh()
}

func (_ *CardList) ViewCard() {

}
