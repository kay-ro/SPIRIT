package trigger

import (
	"fmt"
	"time"
)

var (
	onChange    = func() {}
	recalculate = make(chan bool)
)

func Recalc() {
	recalculate <- true
}

func SetOnChange(f func()) {
	onChange = f
}

func Init() {
	go func() {
		for {
			<-recalculate
			fmt.Println("New Update!! Recalculate", time.Now())
			onChange()
		}
	}()
}
