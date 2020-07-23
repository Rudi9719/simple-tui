package SimpleTui

import (
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
)

// Type for SimpleTUI
// Contains Tui and helper funcs
type SimpleTui struct {
	Tui         *gocui.Gui
	HandleInput func(string) error
	HandleTab   func(string) error
	ListPrint   func(string, ...interface{}) error
	ListTitle   func(string) error
	FeedPrint   func(string, ...interface{}) error
	FeedTitle   func(string) error
	ChatPrint   func(string, ...interface{}) error
	ChatTitle   func(string) error
	InputTitle  func(string) error
}

var (
	g *gocui.Gui
)

func (t SimpleTui) Run() {
	var err error
	g, err = gocui.NewGui(gocui.OutputNormal, false)
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
	t.Tui = g
	t.ListTitle = listTitle
	t.ListPrint = listPrint
	t.FeedTitle = feedTitle
	t.FeedPrint = feedPrint
	t.ChatTitle = chatTitle
	t.ChatPrint = chatPrint
	t.InputTitle = inputTitle
	defer g.Close()
	g.SetManagerFunc(layout)
	if err := t.initKeyBindings(); err != nil {
		fmt.Printf("%+v\n", err)
	}
	if err := g.MainLoop(); err != nil && !gocui.IsQuit(err) {
		fmt.Printf("%+v", err)
	}

}

func layout(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()
	if editView, err := g.SetView("Edit", maxX/2-maxX/3+1, maxY/2, maxX-2, maxY/2+10, 0); err != nil {
		if !gocui.IsUnknownView(err) {
			return err
		}
		editView.Editable = true
		editView.Wrap = true
	}
	if feedView, err := g.SetView("Feed", maxX/2-maxX/3, 0, maxX-1, maxY/5, 0); err != nil {
		if !gocui.IsUnknownView(err) {
			return err
		}
		feedView.Autoscroll = true
		feedView.Wrap = true
	}
	if chatView, err := g.SetView("Chat", maxX/2-maxX/3, maxY/5+1, maxX-1, maxY-5, 0); err != nil {
		if !gocui.IsUnknownView(err) {
			return err
		}
		chatView.Autoscroll = true
		chatView.Wrap = true
	}
	if inputView, err := g.SetView("Input", maxX/2-maxX/3, maxY-4, maxX-1, maxY-1, 0); err != nil {
		if !gocui.IsUnknownView(err) {
			return err
		}
		inputView.Editable = true
		inputView.Wrap = true
		g.Cursor = true
	}
	if listView, err := g.SetView("List", 0, 0, maxX/2-maxX/3-1, maxY-1, 0); err != nil {
		if !gocui.IsUnknownView(err) {
			return err
		}
		listView.Title = "List View"
	}

	return nil
}
func scrollViewUp(v *gocui.View) error {
	return scrollView(v, -1)
}
func scrollViewDown(v *gocui.View) error {
	return scrollView(v, 1)
}
func scrollView(v *gocui.View, delta int) error {
	if v != nil {
		_, y := v.Size()
		ox, oy := v.Origin()
		if oy+delta > strings.Count(v.ViewBuffer(), "\n")-y {
			v.Autoscroll = true
		} else {
			v.Autoscroll = false
			if err := v.SetOrigin(ox, oy+delta); err != nil {
				return err
			}
		}
	}
	return nil
}
func autoScrollView(vn string) error {
	v, err := g.View(vn)
	if err != nil {
		return err
	} else if v != nil {
		v.Autoscroll = true
	}
	return nil
}
func (t SimpleTui) initKeyBindings() error {
	if err := g.SetKeybinding("", gocui.KeyPgup, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			cv, _ := g.View("Chat")
			return scrollViewUp(cv)
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyPgdn, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			cv, _ := g.View("Chat")
			return scrollViewDown(cv)
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			return autoScrollView("Chat")
		}); err != nil {
		return nil
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			input, err := getInputString("Input")
			if err != nil {
				return err
			}
			if input != "" {
				return clearView("Input")
			}
			return gocui.ErrQuit
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("Edit", gocui.KeyCtrlC, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			popupView("Chat")
			popupView("Input")
			clearView("Edit")
			return nil
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("Input", gocui.KeyEnter, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			return t.HandleInput("Input")
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("Input", gocui.KeyTab, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			return t.HandleTab("Input")
		}); err != nil {
		return err
	}
	return nil
}

func setViewTitle(viewName string, title string) {
	g.Update(func(g *gocui.Gui) error {
		updatingView, err := g.View(viewName)
		if err != nil {
			return err
		}
		updatingView.Title = title
		return nil
	})
}
func listTitle(s string) error {
	setViewTitle("List", s)
	return nil
}
func feedTitle(s string) error {
	setViewTitle("Feed", s)
	return nil
}
func chatTitle(s string) error {
	setViewTitle("Chat", s)
	return nil
}
func inputTitle(s string) error {
	setViewTitle("Input", s)
	return nil
}
func getViewTitle(viewName string) string {
	view, err := g.View(viewName)
	if err != nil {
		return ""
	}
	return strings.Split(view.Title, "||")[0]
}

func popupView(viewName string) error {
	_, err := g.SetCurrentView(viewName)
	if err != nil {
		return err
	}
	_, err = g.SetViewOnTop(viewName)
	if err != nil {
		return err
	}
	g.Update(func(g *gocui.Gui) error {
		updatingView, err := g.View(viewName)
		if err != nil {
			return err
		}
		updatingView.MoveCursor(0, 0, true)
		return nil
	})
	return nil
}
func moveCursorToEnd(viewName string) error {
	g.Update(func(g *gocui.Gui) error {
		inputView, err := g.View(viewName)
		if err != nil {
			return err
		}
		inputString, _ := getInputString(viewName)
		stringLen := len(inputString)
		maxX, _ := inputView.Size()
		x := stringLen % maxX
		y := stringLen / maxX
		inputView.SetCursor(0, 0)
		inputView.SetOrigin(0, 0)
		inputView.MoveCursor(x, y, true)
		return nil
	})
	return nil
}
func clearView(viewName string) error {
	g.Update(func(g *gocui.Gui) error {
		inputView, err := g.View(viewName)
		if err != nil {
			return err
		}
		inputView.Clear()
		inputView.SetCursor(0, 0)
		inputView.SetOrigin(0, 0)
		return nil
	})
	return nil
}
func writeToView(viewName string, message string) error {
	g.Update(func(g *gocui.Gui) error {
		updatingView, err := g.View(viewName)
		if err != nil {
			return err
		}
		for _, c := range message {
			updatingView.EditWrite(c)
		}
		return nil
	})
	return nil
}

func printToView(viewName string, message string, a ...interface{}) {
	g.Update(func(g *gocui.Gui) error {
		updatingView, err := g.View(viewName)
		if err != nil {
			return err
		}
		fmt.Fprintf(updatingView, message, a...)
		return nil
	})
}

func listPrint(s string, a ...interface{}) error {
	printToView("List", s, a...)
	return nil
}

func feedPrint(s string, a ...interface{}) error {
	printToView("Feed", s, a...)
	return nil
}
func chatPrint(s string, a ...interface{}) error {
	printToView("Chat", s, a...)
	return nil
}

func getInputString(viewName string) (string, error) {
	inputView, err := g.View(viewName)
	if err != nil {
		return "", err
	}
	retString := inputView.Buffer()
	retString = strings.Replace(retString, "\n", "", -1)
	return retString, err
}
func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func handleInput(viewName string) error {
	clearView(viewName)
	inputString, _ := getInputString(viewName)
	if inputString == "" {
		return nil
	}
	//TODO: Pass this to caller?
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
