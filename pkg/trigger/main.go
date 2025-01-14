package trigger

var (
	onChange    = func() {}
	recalculate = make(chan bool)
)

// trigger a recalculation
//
// this should be used if you want to trigger a recalculation
// after some input fields have been changed
func Recalc() {
	recalculate <- true
}

// sets the onchange function which is getting called after a recalculation trigger
func SetOnChange(f func()) {
	onChange = f
}

// Initializes the channel for notifications
func Init() {
	go func() {
		for {
			<-recalculate
			onChange()
		}
	}()
}
