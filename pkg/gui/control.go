package gui

import (
	"fmt"
	"math"
	"physicsGUI/pkg/gui/helper"
	"physicsGUI/pkg/gui/param"
	"physicsGUI/pkg/minimizer"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	minuit "github.com/empack/minuit2go/pkg"
)

type MinimizerState int

const (
	MinimizerNotStarted = MinimizerState(0x1 << iota)
	MinimizerRunning    = MinimizerState(0x1 << iota)
	MinimizerPaused     = MinimizerState(0x1 << iota)
	MinimizerFinished   = MinimizerState(0x1 << iota)
	MinimizerFailed     = MinimizerState(0x1 << iota)
)

type SharedMinimizerData struct {
	rw       sync.RWMutex
	mnParams *minuit.MnUserParameters
	mFunc    *minimizer.MinuitFunction
	err      error
}

func (controlPanel *MinimizerControlPanel) MinuitUpdateHandler() {
	var migrad *minuit.MnMigrad
	var migrad2 *minuit.MnMigrad
	lastError := math.MaxFloat64

	stateReader := func() MinimizerState {
		controlPanel.rw.RLock()
		defer controlPanel.rw.RUnlock()
		return controlPanel.state
	}

	for {
		switch stateReader() {
		case MinimizerRunning:
			controlPanel.sharedStorage.rw.Lock()
			if migrad == nil {
				// create migrad
				migrad = minuit.NewMnMigradWithParameters(controlPanel.sharedStorage.mFunc, controlPanel.sharedStorage.mnParams)
			}

			res, err := migrad.MinimizeWithMaxfcn(50)

			if err != nil {
				controlPanel.SetStats(err, 0, 0)
				controlPanel.sharedStorage.err = err
				migrad = nil
				migrad2 = nil
				controlPanel.sharedStorage.rw.Unlock()
				controlPanel.Failed(err)
				continue
			}

			if !res.IsValid() {
				if migrad2 == nil {
					migrad2 = minuit.NewMnMigradWithParameterStateStrategy(controlPanel.sharedStorage.mFunc, res.UserState(), minuit.NewMnStrategyWithStra(minuit.PreciseStrategy))
				}
				res, err = migrad2.MinimizeWithMaxfcn(50)
				if err != nil {
					controlPanel.SetStats(err, 0, 0)
					controlPanel.sharedStorage.err = err
					migrad = nil
					migrad2 = nil
					controlPanel.sharedStorage.rw.Unlock()
					controlPanel.Failed(err)
					continue
				}
			}
			controlPanel.SetStats(err, res.Fval(), res.Nfcn())
			_ = controlPanel.sharedStorage.mFunc.UpdateParameters(res.UserParameters().Params())
			controlPanel.sharedStorage.rw.Unlock()
			if res.Fval() == lastError {
				migrad = nil
				migrad2 = nil
				lastError = math.MaxFloat64
				controlPanel.Completed()
			} else {
				lastError = res.Fval()
			}
			continue
		case MinimizerPaused:
			time.Sleep(500 * time.Millisecond) // TODO idle efficient and responsive
			continue
		case MinimizerFinished | MinimizerFailed | MinimizerNotStarted:
			migrad = nil
			migrad2 = nil
			lastError = math.MaxFloat64
			time.Sleep(1000 * time.Millisecond) // TODO idle efficient and responsive
			continue
		}
	}
}

type MinimizerControlPanel struct {
	rw               sync.RWMutex
	state            MinimizerState
	btnPause         *widget.Button
	btnContinue      *widget.Button
	btnStop          *widget.Button
	btnStart         *widget.Button
	lblNCalls        *widget.Label
	lblFVal          *widget.Label
	lblError         *widget.Label
	lblStatus        *widget.Label
	oldMinimizerData []float64
	sharedStorage    *SharedMinimizerData
}

func NewMinimizerControlPanel() *MinimizerControlPanel {
	pnlControl := &MinimizerControlPanel{
		rw:               sync.RWMutex{},
		state:            MinimizerNotStarted,
		btnPause:         nil,
		btnContinue:      nil,
		btnStop:          nil,
		btnStart:         nil,
		lblNCalls:        widget.NewLabel("Calls: -"),
		lblFVal:          widget.NewLabel("FVal: -"),
		lblError:         widget.NewLabel(""),
		lblStatus:        widget.NewLabel("Not Initialized"),
		oldMinimizerData: nil,
		sharedStorage:    &SharedMinimizerData{},
	}
	go pnlControl.MinuitUpdateHandler()
	pnlControl.lblError.Hide()
	pnlControl.btnPause = widget.NewButtonWithIcon("Pause", theme.MediaPauseIcon(), pnlControl.Pause)
	pnlControl.btnPause.Disable()
	pnlControl.btnPause.Hide()
	pnlControl.btnContinue = widget.NewButtonWithIcon("Resume", theme.NavigateNextIcon(), pnlControl.Continue)
	pnlControl.btnContinue.Disable()
	pnlControl.btnContinue.Hide()
	pnlControl.btnStart = widget.NewButtonWithIcon("Start", theme.MediaPlayIcon(), pnlControl.Start)
	pnlControl.btnStop = widget.NewButtonWithIcon("Stop", theme.MediaStopIcon(), pnlControl.Stop)
	pnlControl.btnStop.Disable()
	pnlControl.btnStop.Hide()
	return pnlControl
}

func (controlPanel *MinimizerControlPanel) Widget() fyne.CanvasObject {
	return container.NewHBox(controlPanel.btnStart, controlPanel.btnContinue, controlPanel.btnPause, controlPanel.btnStop, helper.CreateSeparator(), container.NewVBox(container.NewHBox(controlPanel.lblError, controlPanel.lblFVal, controlPanel.lblNCalls), helper.CreateSeparator(), controlPanel.lblStatus))
}

func (controlPanel *MinimizerControlPanel) Pause() {
	if controlPanel.state == MinimizerRunning {
		controlPanel.btnStart.Disable()
		controlPanel.btnStart.Hide()
		controlPanel.btnPause.Disable()
		controlPanel.btnPause.Hide()
		controlPanel.btnContinue.Disable()
		controlPanel.btnContinue.Hide()
		controlPanel.btnStop.Disable()
		controlPanel.btnStop.Hide()

		controlPanel.state = MinimizerPaused

		// this blocks until current cycle is completed
		controlPanel.sharedStorage.rw.Lock()
		controlPanel.sharedStorage.rw.Unlock()

		controlPanel.lblStatus.SetText("Paused")

		controlPanel.btnContinue.Enable()
		controlPanel.btnContinue.Show()
		controlPanel.btnStop.Enable()
		controlPanel.btnStop.Show()

	}
}

func (controlPanel *MinimizerControlPanel) Continue() {
	if controlPanel.state == MinimizerPaused {
		controlPanel.btnStart.Disable()
		controlPanel.btnStart.Hide()
		controlPanel.btnPause.Disable()
		controlPanel.btnPause.Hide()
		controlPanel.btnContinue.Disable()
		controlPanel.btnContinue.Hide()
		controlPanel.btnStop.Disable()
		controlPanel.btnStop.Hide()

		controlPanel.state = MinimizerRunning

		// this blocks until current cycle is completed
		controlPanel.sharedStorage.rw.Lock()
		controlPanel.sharedStorage.rw.Unlock()
		controlPanel.lblStatus.SetText("Running")

		controlPanel.btnPause.Enable()
		controlPanel.btnPause.Show()
		controlPanel.btnStop.Enable()
		controlPanel.btnStop.Show()
	}
}

func (controlPanel *MinimizerControlPanel) Start() {
	if controlPanel.state == MinimizerNotStarted || controlPanel.state == MinimizerFinished || controlPanel.state == MinimizerFailed {
		err := controlPanel.minimizerProblemSetup()
		if err != nil {
			dialog.ShowError(err, MainWindow)
			return
		}
		controlPanel.sharedStorage.rw.RLock()
		controlPanel.oldMinimizerData = controlPanel.sharedStorage.mnParams.Params()
		controlPanel.sharedStorage.rw.RUnlock()

		controlPanel.btnStart.Disable()
		controlPanel.btnStart.Hide()
		controlPanel.btnPause.Disable()
		controlPanel.btnPause.Hide()
		controlPanel.btnContinue.Disable()
		controlPanel.btnContinue.Hide()
		controlPanel.btnStop.Disable()
		controlPanel.btnStop.Hide()

		controlPanel.state = MinimizerPaused

		// this blocks until current cycle is completed
		controlPanel.sharedStorage.rw.Lock()
		controlPanel.sharedStorage.rw.Unlock()
		controlPanel.lblStatus.SetText("Ready")

		controlPanel.Continue()
	}
}

func (controlPanel *MinimizerControlPanel) Stop() {
	controlPanel.btnStart.Disable()
	controlPanel.btnStart.Hide()
	controlPanel.btnPause.Disable()
	controlPanel.btnPause.Hide()
	controlPanel.btnContinue.Disable()
	controlPanel.btnContinue.Hide()
	controlPanel.btnStop.Disable()
	controlPanel.btnStop.Hide()

	controlPanel.Reset()
	controlPanel.SetStats(nil, 0, 0)

	controlPanel.sharedStorage.rw.Lock()
	_ = controlPanel.sharedStorage.mFunc.UpdateParameters(controlPanel.oldMinimizerData)
	controlPanel.sharedStorage.rw.Unlock()
	controlPanel.lblStatus.SetText("Not Initialized")
}

func (controlPanel *MinimizerControlPanel) Reset() {
	controlPanel.btnStart.Disable()
	controlPanel.btnStart.Hide()
	controlPanel.btnPause.Disable()
	controlPanel.btnPause.Hide()
	controlPanel.btnContinue.Disable()
	controlPanel.btnContinue.Hide()
	controlPanel.btnStop.Disable()
	controlPanel.btnStop.Hide()

	controlPanel.state = MinimizerNotStarted

	// this blocks until current cycle is completed
	controlPanel.sharedStorage.rw.Lock()
	controlPanel.sharedStorage.rw.Unlock()
	controlPanel.lblStatus.SetText("Not Initialized")
	controlPanel.btnStart.Enable()
	controlPanel.btnStart.Show()

}

func (controlPanel *MinimizerControlPanel) Completed() {
	controlPanel.Reset()
	dialog.ShowInformation("Minimizer Completed", "Minimizer finished. No further improvements found.", MainWindow)
	controlPanel.state = MinimizerFinished
	// this blocks until current cycle is completed
	controlPanel.sharedStorage.rw.Lock()
	controlPanel.sharedStorage.rw.Unlock()
	controlPanel.lblStatus.SetText("Completed")
}

func (controlPanel *MinimizerControlPanel) Failed(err error) {
	controlPanel.Reset()
	dialog.ShowError(err, MainWindow)
	//TODO ask user if he wants to use data?
	controlPanel.state = MinimizerFailed
	// this blocks until current cycle is completed
	controlPanel.sharedStorage.rw.Lock()
	controlPanel.sharedStorage.rw.Unlock()
	controlPanel.lblStatus.SetText("Failed")
}

func (controlPanel *MinimizerControlPanel) minimize(parameters ...*param.Parameter[float64]) error {
	mnParams := minuit.NewEmptyMnUserParameters()

	var freeToChangeCnt int = 0

	for i, p := range parameters {
		if p == nil {
			return fmt.Errorf("minimizer: parameter %d is nil", i)
		}

		par, err := p.Get()
		if err != nil {
			return err
		}

		id := fmt.Sprintf("p%d", i)

		// if not checked, add as constant parameter
		if !p.IsChecked() {
			mnParams.Add(id, par)
			continue
		}

		min := p.GetRelative("min")
		max := p.GetRelative("max")

		// if min or max is nil, add as free parameter
		if min == nil || max == nil {
			mnParams.AddFree(id, par, 0.1)
			freeToChangeCnt++
			continue
		}

		// if min and max are set, add as limited parameter
		minV, err := min.Get()
		if err != nil {
			return err
		}

		maxV, err := max.Get()
		if err != nil {
			return err
		}

		mnParams.AddLimited(id, par, 0.1, minV, maxV)
		freeToChangeCnt++
	}

	if freeToChangeCnt == 0 {
		return fmt.Errorf("minimizer: No parameter(s) selected to be minimized")
	}

	// create minuit setup
	mFunc := minimizer.NewMinuitFcn(penaltyFunction, parameters)

	controlPanel.sharedStorage.rw.Lock()
	controlPanel.sharedStorage.mFunc = mFunc
	controlPanel.sharedStorage.mnParams = mnParams
	controlPanel.sharedStorage.rw.Unlock()

	return nil
}

func (controlPanel *MinimizerControlPanel) SetStats(err error, fVal float64, nCalls int) {
	if err != nil {
		controlPanel.lblError.SetText(err.Error())
		controlPanel.lblError.Show()
		controlPanel.lblFVal.Hide()
		controlPanel.lblNCalls.Hide()
	} else {
		controlPanel.lblError.SetText("")
		controlPanel.lblError.Hide()
		controlPanel.lblFVal.Show()
		controlPanel.lblNCalls.Show()
	}
	if fVal == 0 {
		controlPanel.lblFVal.SetText(fmt.Sprintf("FVal: -"))
	} else {
		controlPanel.lblFVal.SetText(fmt.Sprintf("FVal: %g", fVal))
	}
	if nCalls == 0 {
		controlPanel.lblNCalls.SetText(fmt.Sprintf("Calls: -"))
	} else {
		controlPanel.lblNCalls.SetText(fmt.Sprintf("Calls: %d", nCalls))
	}
}
