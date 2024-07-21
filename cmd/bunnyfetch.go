package cmd

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var noDistroColor = false
var RootCmd = &cobra.Command{
	Use:   "bunnyfetch",
	Short: "Tiny system info fetch utility.",
	Run: func(cmd *cobra.Command, args []string) {
		bunnyfetch()
	},
}
var r = "\033[0m"
var foot = "\033[31m\"" + r

var Bunny = `            %s
            %s
   (\ /)    %s
   ( . .)   %s
   c(%s)(%s)  %s
	    %s
`

func init() {
	RootCmd.PersistentFlags().BoolVarP(&noDistroColor, "no-distro-color", "d", false, "don't use the color defined in os-release for the OS color")
}

func bunnyfetch() {
	// /etc/os-release should always exist on Linux
	godotenv.Load("/etc/os-release")
	fmt.Printf(
		Bunny,
		titleinf(),
		osinf(),
		kernelinf(),
		shellinf(),
		foot, foot,
		wminf(),
		terminal())
	fmt.Printf("\n         ")
	for i := 0; i < 8; i++ {
		fmt.Printf("\033[4%dm   ", i)
	}
	fmt.Printf(r + "\n         ")
	for i := 0; i < 8; i++ {
		fmt.Printf("\033[10%dm   ", i)
	}
	fmt.Println(r + "\n")
}

func titleinf() string {
	return "\033[31m" + Title() + r
}

func osinf() string {
	// ansi colors of distro (name)? from /etc/os-release
	color := os.Getenv("ANSI_COLOR")
	if color == "" || noDistroColor {
		color = "32"
	}

	return "\033[" + color + "mOS " + r + OS()
}

func kernelinf() string {
	return "\033[33mKernel " + r + Kernel()
}

func shellinf() string {
	return "\033[34mShell " + r + Shell()
}

func wminf() string {
	return "\033[35mWM " + r + WM()
}

func terminal() string {
	return "\033[36mTerminal " + r + Terminal()
}
