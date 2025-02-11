package main

import (
	"log"
	"math"
	"testing"

	minuit "github.com/empack/minuit2go/pkg"
)

// RosenbrockFcn implementiert die Rosenbrock-Funktion für Minuit
type RosenbrockFcn struct{}

func NewRosenbrockFcn() *RosenbrockFcn {
	return &RosenbrockFcn{}
}

// ValueOf implementiert die Rosenbrock-Funktion: f(x,y) = (1-x)² + 100(y-x²)²
func (r *RosenbrockFcn) ValueOf(par []float64) float64 {
	x := par[0]
	y := par[1]

	term1 := math.Pow(1-x, 2)
	term2 := 100 * math.Pow(y-math.Pow(x, 2), 2)

	return term1 + term2
}

func TestRosenbrock(t *testing.T) {
	// Erstelle die Funktion
	theFCN := NewRosenbrockFcn()

	// Parameter Setup
	upar := minuit.NewEmptyMnUserParameters()
	upar.AddFree("x", 10.2, 0.0001)
	upar.AddFree("y", 0.9, 0.0001)

	log.Printf("Initial parameters: %s\n", upar)

	migrad := minuit.NewMnMigradWithParametersStra(theFCN, upar, minuit.PreciseStrategy)
	min, err := migrad.MinimizeWithMaxfcnToler(0, 0.00001)
	if err != nil {
		t.Fatalf("minimize failed with:\n %s\n", err.Error())
	}

	// Falls die erste Minimierung nicht erfolgreich war, versuche es mit höherer Strategie
	if !min.IsValid() {
		println("FM is invalid, try with strategy = 2.")
		migrad2 := minuit.NewMnMigradWithParameterStateStrategy(theFCN, min.UserState(), minuit.NewMnStrategyWithStra(minuit.PreciseStrategy))
		min, err = migrad2.Minimize()
		if err != nil {
			t.Fatalf("minimize failed with:\n %s\n", err.Error())
		}
	}

	// Drucke das Ergebnis
	log.Printf("minimum: %s\n", minuit.MnPrint.ToStringFunctionMinimum(min))

	params := min.UserState().Params()
	if math.Abs(params[0]-1.0) > 1e-3 || math.Abs(params[1]-1.0) > 1e-3 {
		t.Errorf("Minimizer did not find correct minimum. Expected (1,1), got (%f,%f)", params[0], params[1])
	}
}
