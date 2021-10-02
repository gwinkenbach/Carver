package qtui

import (
	"log"

	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

type MenuListener interface {
	DoMenuChoice(menuID uint32)
}

type menuEntry struct {
	ID       uint32
	shortcut gui.QKeySequence__StandardKey
	Label    string
}

type menuEntries struct {
	ID      uint32
	entries []menuEntry
}

type menu struct {
	ID        uint32
	Label     string
	EntriesID uint32
}

type MenuBar struct {
	mainWindow *widgets.QMainWindow
	menuItems  map[uint32]*widgets.QAction
	listener   MenuListener
}

const (
	separatorLabel = "___separator___"

	IDNone = 0

	IDFileMenu    = 1000
	IDNewModel    = 1001
	IDOpenModel   = 1002
	IDSaveModel   = 1003
	IDSaveModelAs = 1004
	IDOpenImage   = 1005

	IDActionMenu = 2000
	IDGenGRBL    = 2001
)

var allMenus = []menuEntries{
	{
		IDFileMenu,
		[]menuEntry{
			{IDNewModel, gui.QKeySequence__New, "New Model"},
			{IDNone, gui.QKeySequence__UnknownKey, separatorLabel},
			{IDOpenModel, gui.QKeySequence__Open, "Open Model..."},
			{IDSaveModel, gui.QKeySequence__Save, "Save Model"},
			{IDSaveModelAs, gui.QKeySequence__SaveAs, "Save Model As..."},
			{IDNone, gui.QKeySequence__UnknownKey, separatorLabel},
			{IDOpenImage, gui.QKeySequence__UnknownKey, "Open image..."},
		},
	},
	{
		IDActionMenu,
		[]menuEntry{
			{IDGenGRBL, gui.QKeySequence__UnknownKey, "Gen GRBL..."},
		},
	},
}

var mainMenuBar = []menu{
	{1, "File", IDFileMenu},
	{2, "Action", IDActionMenu},
}

func CreateMenuBar(win *widgets.QMainWindow) *MenuBar {
	mb := MenuBar{
		mainWindow: win,
		menuItems:  make(map[uint32]*widgets.QAction),
	}

	for _, m := range mainMenuBar {
		qtMenu := win.MenuBar().AddMenu2(m.Label)
		mb.populateMenu(&m, qtMenu)
	}

	return &mb
}

func (mb *MenuBar) EnableMenuItem(menuID uint32, enabled bool) {
	act := mb.menuItems[menuID]
	if act != nil {
		act.SetEnabled(enabled)
	}
}

func (mb *MenuBar) SetMenuListener(listener MenuListener) {
	mb.listener = listener
}

func (mb *MenuBar) populateMenu(m *menu, qm *widgets.QMenu) {
	me := findMenuEntries(m.EntriesID)
	for _, mEntry := range me.entries {
		if mEntry.Label == separatorLabel {
			qm.AddSeparator()
		} else {
			menuID := mEntry.ID
			act := qm.AddAction(mEntry.Label)
			act.ConnectTriggered(func(checked bool) {
				if mb.listener != nil {
					mb.listener.DoMenuChoice(menuID)
				}
			})

			if mEntry.shortcut != gui.QKeySequence__UnknownKey {
				act.SetShortcut(gui.NewQKeySequence5(mEntry.shortcut))
			}

			mb.menuItems[menuID] = act
		}
	}
}

func findMenuEntries(id uint32) *menuEntries {
	for _, me := range allMenus {
		if me.ID == id {
			return &me
		}
	}

	log.Fatalf("Invalid menu-entry ID = %d", id)
	return nil
}
