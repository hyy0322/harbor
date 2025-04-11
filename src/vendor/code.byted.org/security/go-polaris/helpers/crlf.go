package helpers

import "strings"

func FilterCRLFToSpace(input string) string {
	return strings.ReplaceAll(input, "\r\n", "  ")
}
