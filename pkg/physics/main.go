package physics

const (
	ELECTRON_RADIUS = 2.81e-5 // classical electron radius in angstrom

	ZNUMBER  = 150
	QZNUMBER = 500
)

// need more than one? => Copy and rename it. Be careful when to use which axis.
var qzAxis = GetDefaultQZAxis(QZNUMBER)
