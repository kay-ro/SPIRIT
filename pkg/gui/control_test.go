package gui

import (
	"physicsGUI/pkg/function"
	"physicsGUI/pkg/gui/param"
	"runtime"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
)

var once bool = false

func TestSetup(t *testing.T) {
	if once {
		return
	}
	pnlMinimizerUUt := NewMinimizerControlPanel()
	MainWindow = test.NewWindow(pnlMinimizerUUt.Widget())
	registerParams()
	param.GetFloatGroup("eden").GetParam("Eden a").SetCheck(true)
	registerFunctions()
	registerGraphs()
	graphMap["intensity"].AddDataTrack(function.NewEmptyFunction(function.INTERPOLATION_NONE))
	// Set GODEBUG to enable scheduler trace for debugging
	t.Setenv("GODEBUG", "schedtrace=1000")
	// Limit to a single OS thread to control goroutines
	runtime.GOMAXPROCS(1)
	once = true
}

func TestNewMinimizerControlPanel(t *testing.T) {
	TestSetup(t)
	pnlMinimizerUUt := NewMinimizerControlPanel()
	assert.NotNil(t, pnlMinimizerUUt)
	assert.Equal(t, MinimizerNotStarted, pnlMinimizerUUt.state)
	assert.NotNil(t, pnlMinimizerUUt.sharedStorage)
}

func TestMinimizerControlPanel_Start(t *testing.T) {
	TestSetup(t)
	pnlMinimizerUUt := NewMinimizerControlPanel()
	pnlMinimizerUUt.Start()
	assert.Equal(t, MinimizerRunning, pnlMinimizerUUt.state)
}

func TestMinimizerControlPanel_Pause(t *testing.T) {
	TestSetup(t)
	pnlMinimizerUUt := NewMinimizerControlPanel()
	pnlMinimizerUUt.state = MinimizerRunning
	pnlMinimizerUUt.Pause()
	assert.Equal(t, MinimizerPaused, pnlMinimizerUUt.state)
}

func TestMinimizerControlPanel_Continue(t *testing.T) {
	TestSetup(t)
	pnlMinimizerUUt := NewMinimizerControlPanel()
	pnlMinimizerUUt.state = MinimizerPaused
	pnlMinimizerUUt.Continue()
	assert.Equal(t, MinimizerRunning, pnlMinimizerUUt.state)
}

func TestMinimizerControlPanel_Stop(t *testing.T) {
	//t.Skipf("Cant test in with dummy function, because minimizer is instant finished")
	TestSetup(t)
	pnlMinimizerUUt := NewMinimizerControlPanel()
	pnlMinimizerUUt.Start()
	pnlMinimizerUUt.Stop()
	assert.Equal(t, MinimizerNotStarted, pnlMinimizerUUt.state)
}

func TestMinimizerControlPanel_Completed(t *testing.T) {
	TestSetup(t)
	pnlMinimizerUUt := NewMinimizerControlPanel()
	pnlMinimizerUUt.Completed()
	assert.Equal(t, MinimizerFinished, pnlMinimizerUUt.state)
}

func TestMinimizerControlPanel_Failed(t *testing.T) {
	TestSetup(t)
	pnlMinimizerUUt := NewMinimizerControlPanel()
	err := pnlMinimizerUUt.minimize(nil)
	pnlMinimizerUUt.Failed(err)
	assert.Equal(t, MinimizerFailed, pnlMinimizerUUt.state)
}

func TestMinimizerControlPanel_SetStats(t *testing.T) {
	TestSetup(t)
	pnlMinimizerUUt := NewMinimizerControlPanel()
	pnlMinimizerUUt.SetStats(nil, 10.5, 5)
	assert.Equal(t, "FVal: 10.5", pnlMinimizerUUt.lblFVal.Text)
	assert.Equal(t, "Calls: 5", pnlMinimizerUUt.lblNCalls.Text)
}

func TestButtonStart(t *testing.T) {
	TestSetup(t)
	pnlMinimizerUUt := NewMinimizerControlPanel()
	test.Tap(pnlMinimizerUUt.btnStart)
	//time.Sleep(500 * time.Millisecond)
	assert.Equal(t, MinimizerRunning, pnlMinimizerUUt.state)
}

func TestButtonPause(t *testing.T) {
	t.Skipf("Cant test in with dummy function, because minimizer is instant finished and sharedStorage is not set up") //TODO when better id or changed structure fo control.go and main.go
	TestSetup(t)
	pnlMinimizerUUt := NewMinimizerControlPanel()
	pnlMinimizerUUt.Start()
	test.Tap(pnlMinimizerUUt.btnPause)
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, MinimizerPaused, pnlMinimizerUUt.state)
}

func TestButtonContinue(t *testing.T) {
	t.Skipf("Cant test in with dummy function, because minimizer is instant finished and sharedStorage is not set up") //TODO when better id or changed structure fo control.go and main.go
	TestSetup(t)
	pnlMinimizerUUt := NewMinimizerControlPanel()
	pnlMinimizerUUt.state = MinimizerPaused
	test.Tap(pnlMinimizerUUt.btnContinue)
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, MinimizerRunning, pnlMinimizerUUt.state)
}

func TestButtonStop(t *testing.T) {
	t.Skipf("Cant test in with dummy function, because minimizer is instant finished and sharedStorage is not set up") //TODO when better id or changed structure fo control.go and main.go

	TestSetup(t)
	pnlMinimizerUUt := NewMinimizerControlPanel()
	pnlMinimizerUUt.state = MinimizerRunning
	test.Tap(pnlMinimizerUUt.btnStop)
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, MinimizerNotStarted, pnlMinimizerUUt.state)
}

func TestButtonSequence(t *testing.T) {
	t.Skipf("Cant test in with dummy function, because minimizer is instant finished and sharedStorage is not set up") //TODO when better id or changed structure fo control.go and main.go
	TestSetup(t)
	pnlMinimizerUUt := NewMinimizerControlPanel()

	// Versuche Continue zu drücken, sollte ignoriert werden
	test.Tap(pnlMinimizerUUt.btnContinue)
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, MinimizerNotStarted, pnlMinimizerUUt.state)

	// Start drücken, dann Pause drücken
	test.Tap(pnlMinimizerUUt.btnStart)
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, MinimizerRunning, pnlMinimizerUUt.state)
	test.Tap(pnlMinimizerUUt.btnPause)
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, MinimizerPaused, pnlMinimizerUUt.state)
}
