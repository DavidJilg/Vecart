package main

import (
	"embed"
)

//go:embed all:static
var StaticAssets embed.FS

//go:embed LICENSE
var License embed.FS
