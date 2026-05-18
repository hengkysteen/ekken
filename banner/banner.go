package banner

import (
	"fmt"
)

const (
	colorReset = "\033[0m"
	colorBlue  = "\033[38;5;75m"
	colorCyan  = "\033[38;5;81m"
	colorGray  = "\033[38;5;245m"
	colorBold  = "\033[1m"
)

func PrintDev(version, address string) {
	printBanner(version, address, true)
}

func PrintProd(version, address string) {
	printBanner(version, address, false)
}

func printBanner(version, address string, isDev bool) {
	asciiArt := `
  ███████╗██╗  ██╗██╗  ██╗███████╗███╗   ██╗
  ██╔════╝██║ ██╔╝██║ ██╔╝██╔════╝████╗  ██║
  █████╗  █████╔╝ █████╔╝ █████╗  ██╔██╗ ██║
  ██╔══╝  ██╔═██╗ ██╔═██╗ ██╔══╝  ██║╚██╗██║
  ███████╗██║  ██╗██║  ██╗███████╗██║ ╚████║
  ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═══╝`

	fmt.Println(colorBlue + asciiArt + colorReset)
	fmt.Printf("  %sEkken %s%s%s\n", colorBold, colorBlue, version, colorReset)
	fmt.Println()

	if isDev {
		fmt.Printf("  %s● %sMode:    %sDevelopment%s\n", colorBlue, colorBold, colorReset, colorReset)
		fmt.Println()
		fmt.Printf("  %s➜ %s API: %shttp://%s%s\n", colorCyan, colorBold, colorReset, address, colorReset)
		fmt.Printf("  %s➜ %s UI:  %scd ui && npm run dev%s\n", colorCyan, colorBold, colorGray, colorReset)
	} else {
		fmt.Printf("  %s➜ %s URL: %shttp://%s%s\n", colorCyan, colorBold, colorReset, address, colorReset)
	}
	fmt.Println()
}
