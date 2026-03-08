package webassets

import (
	"embed"
	"io/fs"
)

//go:embed index.html styles.css app.js downloads/*
var embedded embed.FS

func Files() fs.FS {
	return embedded
}
