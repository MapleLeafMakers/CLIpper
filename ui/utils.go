package ui

import (
	"errors"
	"github.com/charmbracelet/lipgloss"
	"slices"
	"strings"
)

var borderTypeMap = map[string]lipgloss.Border{
	"normal":         lipgloss.NormalBorder(),
	"rounded":        lipgloss.RoundedBorder(),
	"block":          lipgloss.BlockBorder(),
	"outerhalfblock": lipgloss.OuterHalfBlockBorder(),
	"innerhalfblock": lipgloss.InnerHalfBlockBorder(),
	"thick":          lipgloss.ThickBorder(),
	"double":         lipgloss.DoubleBorder(),
	"hidden":         lipgloss.HiddenBorder(),
}

func parseBool(input string) (bool, error) {
	s := strings.ToLower(input)
	if slices.Contains([]string{"no", "n", "0", "false", "f", "off"}, s) {
		return false, nil
	} else if slices.Contains([]string{"yes", "y", "1", "true", "t", "on"}, s) {
		return true, nil
	}
	return false, errors.New("Invalid boolean literal: " + input)
}

func parseBorderType(input string) (lipgloss.Border, error) {
	strings.ToLower(input)
	bt, ok := borderTypeMap[strings.ToLower(input)]
	if !ok {
		return lipgloss.NormalBorder(), errors.New("Invalid border type: " + input)
	}
	return bt, nil
}

func parseColor(input string) (lipgloss.TerminalColor, error) {
	switch strings.ToLower(input) {
	case "none", "null", "nil", "unset", "false":
		return lipgloss.Color(""), nil
	}
	return lipgloss.Color(input), nil
}

func parseTimestampFormat(input string) (string, error) {
	return input, nil
}
