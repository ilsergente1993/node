package dashboard

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	_ "github.com/gizak/termui/v3"
	_ "github.com/gizak/termui/v3/widgets"
)

var clear map[string]func() //create a map for storing clear funcs

func init() {
	clear = make(map[string]func()) //Initialize it
	clear["linux"] = func() {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func CallClear() {
	value, ok := clear[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok { //if we defined a clear func for that platform:
		value() //we execute it
	} else { //unsupported platform
		panic("Your platform is unsupported! I can't clear terminal screen :(")
	}
}

//determinates the best Unit [Bytes | KB | MB | GB] for the passed val
func ToUnit(bytes uint64) string {
	switch {
	case bytes < 1000:
		return strconv.FormatUint(bytes, 10) + " Bytes"
	case bytes < 1000000:
		return strconv.FormatUint(bytes/1000, 10) + " KB"
	case bytes < 1000000000:
		return strconv.FormatUint(bytes/1000000, 10) + " MB"
	case bytes < 1000000000000:
		return strconv.FormatUint(bytes/1000000000, 10) + " GB"
	default:
		return "Toooo big"

	}
}
func Line() string {
	l := ""
	for i := 0; i < WINWIDTH; i++ {
		l += "_"
	}
	return l
}
