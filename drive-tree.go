package main

import (
	"drive-tree/internal/web"
	"flag"
	"fmt"
	"os"
)

func main() {

	addr := flag.String("addr", "127.0.0.1:8080", "address the web interface listens on (use :8080 inside Docker)")
	noBrowser := flag.Bool("no-browser", false, "don't open the browser automatically")
	flag.Usage = usage
	flag.Parse()

	// "scraper" and "web" used to be separate steps, now a single run does both
	if arg := flag.Arg(0); arg != "" && arg != "scraper" && arg != "web" {
		usage()
		os.Exit(2)
	}

	web.Run(*addr, !*noBrowser)
}

func usage() {

	fmt.Println("Usage:")
	fmt.Println("\tdrive-tree [flags] --> sign in, scan your files and view them in the browser")
	fmt.Println()
	fmt.Println("Flags:")
	flag.PrintDefaults()

}
