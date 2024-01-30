package cmdinput

import (
	"github.com/MapleLeafMakers/tview"
	"github.com/gdamore/tcell/v2"
	"strconv"
	"strings"
	"sync"
)

const (
	AutocompletedNavigate = iota // The user navigated the autocomplete list (using the errow keys).
	AutocompletedTab             // The user selected an autocomplete entry using the tab key.
	AutocompletedEnter           // The user selected an autocomplete entry using the enter key.
	AutocompletedClick           // The user selected an autocomplete entry by clicking the mouse button on it.
)

// Predefined InputField acceptance functions.
var (
	// InputFieldInteger accepts integers.
	InputFieldInteger = func(text string, ch rune) bool {
		if text == "-" {
			return true
		}
		_, err := strconv.Atoi(text)
		return err == nil
	}

	// InputFieldFloat accepts floating-point numbers.
	InputFieldFloat = func(text string, ch rune) bool {
		if text == "-" || text == "." || text == "-." {
			return true
		}
		_, err := strconv.ParseFloat(text, 64)
		return err == nil
	}

	// InputFieldMaxLength returns an input field accept handler which accepts
	// input strings up to a given length. Use it like this:
	//
	//   inputField.SetAcceptanceFunc(InputFieldMaxLength(10)) // Accept up to 10 characters.
	InputFieldMaxLength = func(maxLength int) func(text string, ch rune) bool {
		return func(text string, ch rune) bool {
			return len([]rune(text)) <= maxLength
		}
	}
)

type Suggestion struct {
	Text string
	Help string
}

// InputField is a one-line box into which the user can enter text. Use
// [InputField.SetAcceptanceFunc] to accept or reject input,
// [InputField.SetChangedFunc] to listen for changes, and
// [InputField.SetMaskCharacter] to hide input from onlookers (e.g. for password
// input).
//
// The input field also has an optional autocomplete feature. It is initialized
// by the [InputField.SetAutocompleteFunc] function. For more control over the
// autocomplete drop-down's behavior, you can also set the
// [InputField.SetAutocompletedFunc].
//
// Navigation and editing is the same as for a [TextArea], with the following
// exceptions:
//
//   - Tab, BackTab, Enter, Escape: Finish editing.
//
// If autocomplete functionality is configured:
//
//   - Down arrow: Open the autocomplete drop-down.
//   - Tab, Enter: Select the current autocomplete entry.
//
// See https://github.com/rivo/tview/wiki/InputField for an example.
type InputField struct {
	*tview.Box

	// The text area providing the core functionality of the input field.
	textArea *tview.TextArea

	// The screen width of the input area. A value of 0 means extend as much as
	// possible.
	fieldWidth int

	// An optional autocomplete function which receives the current text of the
	// input field and returns a slice of strings to be displayed in a drop-down
	// selection.
	autocomplete func(text string, cursorPos int) ([]Suggestion, int)

	// The List object which shows the selectable autocomplete entries. If not
	// nil, the list's main texts represent the current autocomplete entries.
	autocompleteListOffset int
	autocompleteList       *tview.List
	autocompleteListMutex  sync.Mutex

	// The styles of the autocomplete entries.
	autocompleteStyles struct {
		main       tcell.Style
		selected   tcell.Style
		background tcell.Color
		help       tcell.Style
	}

	commandHistory      []string
	commandHistoryIndex int

	// An optional function which is called when the user selects an
	// autocomplete entry. The text and index of the selected entry (within the
	// list) is provided, as well as the user action causing the selection (one
	// of the "Autocompleted" values). The function should return true if the
	// autocomplete list should be closed. If nil, the input field will be
	// updated automatically when the user navigates the autocomplete list.
	autocompleted func(text string, index int, source int) bool

	// An optional function which may reject the last character that was entered.
	accept func(text string, ch rune) bool

	// An optional function which is called when the input has changed.
	changed func(text string)

	// An optional function which is called when the user indicated that they
	// are done entering text. The key which was pressed is provided (tab,
	// shift-tab, enter, or escape).
	done func(tcell.Key)

	// A callback function set by the Form class and called when the user leaves
	// this form item.
	finished func(tcell.Key)
}

// NewInputField returns a new input field.
func NewInputField() *InputField {

	i := &InputField{
		Box:      tview.NewBox(),
		textArea: tview.NewTextArea().SetWrap(false),
	}

	i.textArea.SetChangedFunc(func() {
		if i.changed != nil {
			i.changed(i.textArea.GetText())
		}
	})

	i.textArea.SetTextStyle(tcell.StyleDefault.Background(tview.Styles.ContrastBackgroundColor).Foreground(tview.Styles.PrimaryTextColor))
	i.textArea.SetPlaceholderStyle(tcell.StyleDefault.Background(tview.Styles.ContrastBackgroundColor).Foreground(tview.Styles.ContrastSecondaryTextColor))
	i.autocompleteStyles.main = tcell.StyleDefault.Foreground(tview.Styles.PrimitiveBackgroundColor)
	i.autocompleteStyles.selected = tcell.StyleDefault.Background(tview.Styles.PrimaryTextColor).Foreground(tview.Styles.PrimitiveBackgroundColor)
	i.autocompleteStyles.background = tview.Styles.MoreContrastBackgroundColor

	commandHistory := make([]string, 0, 200)
	commandHistory = append(commandHistory, "")
	i.commandHistory = commandHistory
	i.commandHistoryIndex = 0
	return i
}

// SetText sets the current text of the input field. This can be undone by the
// user. Calling this function will also trigger a "changed" event.
func (i *InputField) SetText(text string) *InputField {
	i.textArea.Replace(0, i.textArea.GetTextLength(), text)
	return i
}

// GetText returns the current text of the input field.
func (i *InputField) GetText() string {
	return i.textArea.GetText()
}

// SetLabel sets the text to be displayed before the input area.
func (i *InputField) SetLabel(label string) *InputField {
	i.textArea.SetLabel(label)
	return i
}

// GetLabel returns the text to be displayed before the input area.
func (i *InputField) GetLabel() string {
	return i.textArea.GetLabel()
}

// SetLabelWidth sets the screen width of the label. A value of 0 will cause the
// primitive to use the width of the label string.
func (i *InputField) SetLabelWidth(width int) *InputField {
	i.textArea.SetLabelWidth(width)
	return i
}

// SetPlaceholder sets the text to be displayed when the input text is empty.
func (i *InputField) SetPlaceholder(text string) *InputField {
	i.textArea.SetPlaceholder(text)
	return i
}

// SetLabelColor sets the text color of the label.
func (i *InputField) SetLabelColor(color tcell.Color) *InputField {
	i.textArea.SetLabelStyle(i.textArea.GetLabelStyle().Foreground(color))
	return i
}

// SetLabelStyle sets the style of the label.
func (i *InputField) SetLabelStyle(style tcell.Style) *InputField {
	i.textArea.SetLabelStyle(style)
	return i
}

// GetLabelStyle returns the style of the label.
func (i *InputField) GetLabelStyle() tcell.Style {
	return i.textArea.GetLabelStyle()
}

// SetFieldBackgroundColor sets the background color of the input area.
func (i *InputField) SetFieldBackgroundColor(color tcell.Color) *InputField {
	i.textArea.SetTextStyle(i.textArea.GetTextStyle().Background(color))
	return i
}

// SetFieldTextColor sets the text color of the input area.
func (i *InputField) SetFieldTextColor(color tcell.Color) *InputField {
	i.textArea.SetTextStyle(i.textArea.GetTextStyle().Foreground(color))
	return i
}

// SetFieldStyle sets the style of the input area (when no placeholder is
// shown).
func (i *InputField) SetFieldStyle(style tcell.Style) *InputField {
	i.textArea.SetTextStyle(style)
	return i
}

// GetFieldStyle returns the style of the input area (when no placeholder is
// shown).
func (i *InputField) GetFieldStyle() tcell.Style {
	return i.textArea.GetTextStyle()
}

// SetPlaceholderTextColor sets the text color of placeholder text.
func (i *InputField) SetPlaceholderTextColor(color tcell.Color) *InputField {
	i.textArea.SetPlaceholderStyle(i.textArea.GetPlaceholderStyle().Foreground(color))
	return i
}

// SetPlaceholderStyle sets the style of the input area (when a placeholder is
// shown).
func (i *InputField) SetPlaceholderStyle(style tcell.Style) *InputField {
	i.textArea.SetPlaceholderStyle(style)
	return i
}

// GetPlaceholderStyle returns the style of the input area (when a placeholder
// is shown).
func (i *InputField) GetPlaceholderStyle() tcell.Style {
	return i.textArea.GetPlaceholderStyle()
}

// SetAutocompleteStyles sets the colors and style of the autocomplete entries.
// For details, see List.SetMainTextStyle(), List.SetSelectedStyle(), and
// Box.SetBackgroundColor().
func (i *InputField) SetAutocompleteStyles(background tcell.Color, main, selected tcell.Style, help tcell.Style) *InputField {
	i.autocompleteStyles.background = background
	i.autocompleteStyles.main = main
	i.autocompleteStyles.selected = selected
	i.autocompleteStyles.help = help
	return i
}

// SetFormAttributes sets attributes shared by all form items.
func (i *InputField) SetFormAttributes(labelWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) tview.FormItem {
	i.textArea.SetFormAttributes(labelWidth, labelColor, bgColor, fieldTextColor, fieldBgColor)
	return i
}

// SetFieldWidth sets the screen width of the input area. A value of 0 means
// extend as much as possible.
func (i *InputField) SetFieldWidth(width int) *InputField {
	i.fieldWidth = width
	return i
}

// GetFieldWidth returns this primitive's field width.
func (i *InputField) GetFieldWidth() int {
	return i.fieldWidth
}

// GetFieldHeight returns this primitive's field height.
func (i *InputField) GetFieldHeight() int {
	return 1
}

// SetDisabled sets whether or not the item is disabled / read-only.
func (i *InputField) SetDisabled(disabled bool) tview.FormItem {
	i.textArea.SetDisabled(disabled)
	if i.finished != nil {
		i.finished(-1)
	}
	return i
}

// SetAutocompleteFunc sets an autocomplete callback function which may return
// strings to be selected from a drop-down based on the current text of the
// input field. The drop-down appears only if len(entries) > 0. The callback is
// invoked in this function and whenever the current text changes or when
// [InputField.Autocomplete] is called. Entries are cleared when the user
// selects an entry or presses Escape.
func (i *InputField) SetAutocompleteFunc(callback func(currentText string, cursorPos int) (entries []Suggestion, menuOffset int)) *InputField {
	i.autocomplete = callback
	i.Autocomplete()
	return i
}

// SetAutocompletedFunc sets a callback function which is invoked when the user
// selects an entry from the autocomplete drop-down list. The function is passed
// the text of the selected entry (stripped of any style tags), the index of the
// entry, and the user action that caused the selection, for example
// [AutocompletedNavigate]. It returns true if the autocomplete drop-down should
// be closed after the callback returns or false if it should remain open, in
// which case [InputField.Autocomplete] is called to update the drop-down's
// contents.
//
// If no such callback is set (or nil is provided), the input field will be
// updated with the selection any time the user navigates the autocomplete
// drop-down list. So this function essentially gives you more control over the
// autocomplete functionality.
func (i *InputField) SetAutocompletedFunc(autocompleted func(text string, index int, source int) bool) *InputField {
	i.autocompleted = autocompleted
	return i
}

// Autocomplete invokes the autocomplete callback (if there is one, see
// [InputField.SetAutocompleteFunc]). If the length of the returned autocomplete
// entries slice is greater than 0, the input field will present the user with a
// corresponding drop-down list the next time the input field is drawn.
//
// It is safe to call this function from any goroutine. Note that the input
// field is not redrawn automatically unless called from the main goroutine
// (e.g. in response to events).
func (i *InputField) Autocomplete() *InputField {
	i.autocompleteListMutex.Lock()
	defer i.autocompleteListMutex.Unlock()
	if i.autocomplete == nil {
		return i
	}

	// Do we have any autocomplete entries?
	text := i.textArea.GetText()
	_, cursorPos, _, _ := i.textArea.GetCursor()
	entries, menuOffset := i.autocomplete(text, cursorPos)
	if len(entries) == 0 {
		// No entries, no list.
		i.autocompleteList = nil
		return i
	}

	// Make a list if we have none.
	if i.autocompleteList == nil {
		i.autocompleteList = tview.NewList()
		i.autocompleteList.ShowSecondaryText(false).
			SetMainTextStyle(i.autocompleteStyles.main).
			SetSelectedStyle(i.autocompleteStyles.selected).
			SetHighlightFullLine(true).
			SetBackgroundColor(i.autocompleteStyles.background)
	}
	// store the menuOffset returned from the tab completer
	i.autocompleteListOffset = menuOffset

	// Fill it with the entries.
	currentEntry := -1
	suffixLength := 9999 // I'm just waiting for the day somebody opens an issue with this number being too small.
	i.autocompleteList.Clear()
	for index, entry := range entries {
		txt := entry.Text
		if entry.Help != "" {
			fg, bg, _ := i.autocompleteStyles.help.Decompose()
			styleTag := "[" + fg.Name(true) + ":" + bg.Name(true) + ":i]"
			txt += styleTag + " - " + entry.Help
		}
		i.autocompleteList.AddItem(txt, entry.Text, 0, nil)
		if strings.HasPrefix(entry.Text, text) && len(entry.Text)-len(text) < suffixLength {
			currentEntry = index
			suffixLength = len(text) - len(entry.Text)
		}
	}

	// Set the selection if we have one.
	if currentEntry >= 0 {
		i.autocompleteList.SetCurrentItem(currentEntry)
	}

	return i
}

// SetAcceptanceFunc sets a handler which may reject the last character that was
// entered, by returning false. The handler receives the text as it would be
// after the change and the last character entered. If the handler is nil, all
// input is accepted. The function is only called when a single rune is inserted
// at the current cursor position.
//
// This package defines a number of variables prefixed with InputField which may
// be used for common input (e.g. numbers, maximum text length). See for example
// [InputFieldInteger].
func (i *InputField) SetAcceptanceFunc(handler func(textToCheck string, lastChar rune) bool) *InputField {
	i.accept = handler
	return i
}

// SetChangedFunc sets a handler which is called whenever the text of the input
// field has changed. It receives the current text (after the change).
func (i *InputField) SetChangedFunc(handler func(text string)) *InputField {
	i.changed = handler
	return i
}

// SetDoneFunc sets a handler which is called when the user is done entering
// text. The callback function is provided with the key that was pressed, which
// is one of the following:
//
//   - KeyEnter: Done entering text.
//   - KeyEscape: Abort text input.
//   - KeyTab: Move to the next field.
//   - KeyBacktab: Move to the previous field.
func (i *InputField) SetDoneFunc(handler func(key tcell.Key)) *InputField {
	i.done = handler
	return i
}

// SetFinishedFunc sets a callback invoked when the user leaves this form item.
func (i *InputField) SetFinishedFunc(handler func(key tcell.Key)) tview.FormItem {
	i.finished = handler
	return i
}

// Focus is called when this primitive receives focus.
func (i *InputField) Focus(delegate func(p tview.Primitive)) {
	// If we're part of a form and this item is disabled, there's nothing the
	// user can do here so we're finished.
	if i.finished != nil && i.textArea.GetDisabled() {
		i.finished(-1)
		return
	}
	i.textArea.Focus(delegate)
	i.Box.Focus(delegate)
}

// HasFocus returns whether or not this primitive has focus.
func (i *InputField) HasFocus() bool {
	return i.textArea.HasFocus() || i.Box.HasFocus()
}

// Blur is called when this primitive loses focus.
func (i *InputField) Blur() {
	i.textArea.Blur()
	i.Box.Blur()
	i.autocompleteList = nil // Hide the autocomplete drop-down.
}

// Draw draws this primitive onto the screen.
func (i *InputField) Draw(screen tcell.Screen) {
	i.Box.DrawForSubclass(screen, i)

	// Prepare
	x, y, width, height := i.GetInnerRect()
	if height < 1 || width < 1 {
		return
	}

	// Resize text area.
	labelWidth := i.textArea.GetLabelWidth()
	if labelWidth == 0 {
		labelWidth = tview.TaggedStringWidth(i.textArea.GetLabel())
	}
	fieldWidth := i.fieldWidth
	if fieldWidth == 0 {
		fieldWidth = width - labelWidth
	}
	i.textArea.SetRect(x, y, labelWidth+fieldWidth, 1)
	// i.textArea.setMinCursorPadding(fieldWidth-1, 1)
	// Draw text area.
	i.textArea.SetHasFocus(i.HasFocus()) // Force cursor positioning.
	i.textArea.Draw(screen)

	// Draw autocomplete list.
	i.autocompleteListMutex.Lock()
	defer i.autocompleteListMutex.Unlock()
	if i.autocompleteList != nil {
		// How much space do we need?
		lheight := i.autocompleteList.GetItemCount()
		lwidth := 0
		for index := 0; index < lheight; index++ {
			entry, _ := i.autocompleteList.GetItemText(index)
			width := tview.TaggedStringWidth(entry)
			if width > lwidth {
				lwidth = width
			}
		}

		// We prefer to drop down but if there is no space, maybe drop up?
		lx := x + labelWidth + i.autocompleteListOffset
		ly := y + 1
		_, sheight := screen.Size()
		if ly+lheight >= sheight && ly-2 > lheight-ly {
			ly = y - lheight
			if ly < 0 {
				ly = 0
			}
		}
		if ly+lheight >= sheight {
			lheight = sheight - ly
		}
		i.autocompleteList.SetRect(lx, ly, lwidth, lheight)
		i.autocompleteList.Draw(screen)
	}
}

// InputHandler returns the handler for this primitive.
func (i *InputField) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return i.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if i.textArea.GetDisabled() {
			return
		}

		// Trigger changed events.
		var skipAutocomplete bool
		currentText := i.textArea.GetText()
		defer func() {
			newText := i.textArea.GetText()
			if newText != currentText {
				if !skipAutocomplete {
					i.Autocomplete()
				}
				if i.changed != nil {
					i.changed(newText)
				}
			}
		}()

		// If we have an autocomplete list, there are certain keys we will
		// forward to it.
		i.autocompleteListMutex.Lock()
		defer i.autocompleteListMutex.Unlock()
		if i.autocompleteList != nil {
			i.autocompleteList.SetChangedFunc(nil)
			i.autocompleteList.SetSelectedFunc(nil)
			switch key := event.Key(); key {
			case tcell.KeyEscape: // Close the list.
				i.autocompleteList = nil
				return
			case tcell.KeyEnter, tcell.KeyTab: // Intentional selection.
				index := i.autocompleteList.GetCurrentItem()
				_, text := i.autocompleteList.GetItemText(index)
				if i.autocompleted != nil {
					source := AutocompletedEnter
					if key == tcell.KeyTab {
						source = AutocompletedTab
					}
					if i.autocompleted(stripTags(text), index, source) {
						i.autocompleteList = nil
						currentText = i.GetText()
						i.updateCurrentHistoryEntry()
					}
				} else {
					i.SetText(text)
					skipAutocomplete = true
					i.autocompleteList = nil
				}
				return
			case tcell.KeyDown, tcell.KeyUp, tcell.KeyPgDn, tcell.KeyPgUp:
				i.autocompleteList.SetChangedFunc(func(index int, text, secondaryText string, shortcut rune) {
					text = stripTags(text)
					if i.autocompleted != nil {
						if i.autocompleted(text, index, AutocompletedNavigate) {
							i.autocompleteList = nil
							currentText = i.GetText()
						}
					} else {
						i.SetText(text)
						currentText = stripTags(text) // We want to keep the autocomplete list open and unchanged.
					}
				})
				i.autocompleteList.InputHandler()(event, setFocus)
				return
			}
		}

		// Finish up.
		finish := func(key tcell.Key) {
			if i.done != nil {
				i.done(key)
			}
			if i.finished != nil {
				i.finished(key)
			}
		}

		// Process special key events for the input field.
		switch key := event.Key(); key {
		case tcell.KeyDown:
			i.autocompleteListMutex.Unlock() // We're still holding a lock.
			//i.Autocomplete()
			i.autocompleteListMutex.Lock()
			i.HistoryDown()
			skipAutocomplete = true

		case tcell.KeyUp:
			i.HistoryUp()
			skipAutocomplete = true

		case tcell.KeyEnter, tcell.KeyEscape, tcell.KeyTab, tcell.KeyBacktab:
			finish(key)
		case tcell.KeyRune:
			if event.Modifiers()&tcell.ModAlt == 0 && i.accept != nil {
				// Check if this rune is accepted.
				_, cursPos, _, _ := i.textArea.GetCursor()
				txt := i.textArea.GetText()
				textBefore := txt[:cursPos]
				textAfter := txt[cursPos:]
				r := event.Rune()
				if !i.accept(textBefore+string(r)+textAfter, r) {
					return
				}
			}
			fallthrough
		default:
			// Forward other key events to the text area.
			i.textArea.InputHandler()(event, setFocus)
			i.updateCurrentHistoryEntry()
		}
		//i.updateCurrentHistoryEntry()
	})
}

// MouseHandler returns the mouse handler for this primitive.
func (i *InputField) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return i.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if i.textArea.GetDisabled() {
			return false, nil
		}

		var skipAutocomplete bool
		currentText := i.GetText()
		defer func() {
			newText := i.GetText()
			if newText != currentText {
				if !skipAutocomplete {
					i.Autocomplete()
				}
				if i.changed != nil {
					i.changed(newText)
				}
			}
		}()

		// If we have an autocomplete list, forward the mouse event to it.
		i.autocompleteListMutex.Lock()
		defer i.autocompleteListMutex.Unlock()
		if i.autocompleteList != nil {
			i.autocompleteList.SetChangedFunc(nil)
			i.autocompleteList.SetSelectedFunc(func(index int, text, secondaryText string, shortcut rune) {
				text = stripTags(text)
				if i.autocompleted != nil {
					if i.autocompleted(text, index, AutocompletedClick) {
						i.autocompleteList = nil
						currentText = i.GetText()
					}
					return
				}
				i.SetText(text)
				skipAutocomplete = true
				i.autocompleteList = nil
			})
			if consumed, _ = i.autocompleteList.MouseHandler()(action, event, setFocus); consumed {
				setFocus(i)
				return
			}
		}

		// Is mouse event within the input field?
		x, y := event.Position()
		if !i.InRect(x, y) {
			return false, nil
		}

		// Forward mouse event to the text area.
		consumed, capture = i.textArea.MouseHandler()(action, event, setFocus)

		return
	})
}

func (i *InputField) GetCursor() int {
	_, c, _, _ := i.textArea.GetCursor()
	return c
}

func (i *InputField) SetCursor(cursorPos int) {
	i.textArea.Select(cursorPos, cursorPos)
}

func (i *InputField) Clear() {
	i.textArea.SetText("", true)
}

func (i *InputField) HistoryUp() {
	i.nextHistory(-1)
}

func (i *InputField) HistoryDown() {
	i.nextHistory(1)
}

func (i *InputField) nextHistory(dir int) {
	idx := (i.commandHistoryIndex + dir)
	if idx < 0 {
		idx = 0
	} else if idx >= len(i.commandHistory) {
		idx = len(i.commandHistory) - 1
	}
	i.commandHistoryIndex = idx
	i.textArea.SetText(i.commandHistory[i.commandHistoryIndex], true)
}

func (i *InputField) NewCommand() {
	if i.commandHistoryIndex != len(i.commandHistory)-1 {
		i.commandHistory[len(i.commandHistory)-1] = i.GetText()
	}

	i.commandHistory = append(i.commandHistory, "")
	i.commandHistoryIndex = len(i.commandHistory) - 1
	i.Clear()
}

func (i *InputField) updateCurrentHistoryEntry() {
	if i.commandHistoryIndex == len(i.commandHistory)-1 {
		i.commandHistory[i.commandHistoryIndex] = i.GetText()
	}
}

func (i *InputField) GetFocusState() map[string]bool {

	return map[string]bool{
		"Input":          i.Box.HasFocus(),
		"Input.textArea": i.textArea.HasFocus(),
	}
}
