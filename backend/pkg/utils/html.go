package utils

import (
	"fmt"
	"regexp"
)

func InjectBase(htmlContent string, base string) string {
	baseScript := fmt.Sprintf(`<script>window.__dynamic_base__='%s';</script>`, base)
	re := regexp.MustCompile(`<head>`)
	return re.ReplaceAllString(htmlContent, "<head>\n    "+baseScript)
}
