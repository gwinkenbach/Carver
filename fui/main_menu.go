package fui

import (
	"fyne.io/fyne/v2"
)

type MainMenu struct {
	menus     []*fyne.Menu
	menuItems map[string]*fyne.MenuItem
	realized  bool

	onMenuSelected func(menuItemTag string)
}

func NewMainMenu(itemSelected func(menuItemTag string)) *MainMenu {
	return &MainMenu{
		menus:          make([]*fyne.Menu, 0, 8),
		menuItems:      make(map[string]*fyne.MenuItem),
		onMenuSelected: itemSelected,
	}
}

func (mm *MainMenu) AddMenu(label string) {
	if mm.findMenuByLabel(label) == nil {
		mm.menus = append(mm.menus, fyne.NewMenu(label))
	}
}

func (mm *MainMenu) AddMenuItem(
	toMenuLabel string,
	menuItemTag string,
	menuItemLabel string,
	disabled bool) {

	menu := mm.findMenuByLabel(toMenuLabel)
	if menu == nil {
		fyne.LogError("No menu with label "+toMenuLabel, nil)
		return
	}

	_, ok := mm.menuItems[menuItemTag]
	if ok {
		fyne.LogError("Menu item with tag "+menuItemTag+" already exist", nil)
		return
	}

	menuItem := fyne.NewMenuItem(menuItemLabel, func() {
		if mm.onMenuSelected != nil {
			mm.onMenuSelected(menuItemTag)
		}
	})
	menuItem.Disabled = disabled

	menu.Items = append(menu.Items, menuItem)
	mm.menuItems[menuItemTag] = menuItem
}

func (mm *MainMenu) AddSeparator(toMenuLabel string) {
	menu := mm.findMenuByLabel(toMenuLabel)
	if menu == nil {
		fyne.LogError("No menu with label "+toMenuLabel, nil)
		return
	}

	menuItem := fyne.NewMenuItemSeparator()
	menu.Items = append(menu.Items, menuItem)
}

func (mm *MainMenu) SetMenuItemEnabled(tag string, enabled bool) {
	menuItem, ok := mm.menuItems[tag]
	if !ok {
		fyne.LogError("No menu item with tag "+tag, nil)
		return
	}

	menuItem.Disabled = !enabled
}

func (mm *MainMenu) Realize(w fyne.Window) {
	if !mm.realized {
		mainMenu := fyne.NewMainMenu(mm.menus...)
		w.SetMainMenu(mainMenu)
		mm.realized = true
	}
}

func (mm *MainMenu) findMenuByLabel(label string) *fyne.Menu {
	for _, m := range mm.menus {
		if m.Label == label {
			return m
		}
	}

	return nil
}
