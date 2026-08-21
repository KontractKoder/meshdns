package web

import "embed"

var (
	//go:embed index.html
	indexFS embed.FS
)

func HTML() ([]byte, error) {
	return indexFS.ReadFile("index.html")
}
