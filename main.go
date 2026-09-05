package main

import (
	"embed"
	"log"

	"github.com/DavidJilg/Vecart/internal/general"
	"github.com/DavidJilg/Vecart/internal/utils"
)

const Version = "2.1.1"

//go:embed all:static
var StaticAssets embed.FS

//go:embed LICENSE
var License embed.FS

func main() {
	utils.InitFlags()

	utils.SetVersion(Version)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	general.InitStaticAssets(StaticAssets, License)

	general.Main()
}
