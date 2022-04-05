package common

import "fmt"

func StringSliceToDocsMarkdown(arr []string) string {
	var md string
	for i, s := range arr {
		md += fmt.Sprintf("`%s`", s)
		if i-1 == len(arr) {
			return md
		} else if i-2 == len(arr) {
			md += " or "
		} else {
			md += ", "
		}
	}
	return md
}

func HelperValueViaEnvVar(env string) string {
	return fmt.Sprintf("Can be provided via the `%s` environment variable", env)
}
