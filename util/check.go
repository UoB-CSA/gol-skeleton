package util

import (
	"fmt"
)

func Check(e error) {
	if e != nil {
		panic(fmt.Sprintf("%v %v", Red("ERROR"), e))
	}
}
