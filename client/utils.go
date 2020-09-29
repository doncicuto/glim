package client

import "fmt"

func truncate(text string, length int) string {
	if len(text) > length {
		format := fmt.Sprintf("%%.%ds...", length-3)
		return fmt.Sprintf(format, text)
	}
	return text
}
