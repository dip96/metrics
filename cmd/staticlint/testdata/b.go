package a

import "os"

func f() {
	os.Exit(1) // допустимо, так как функция не main
}
