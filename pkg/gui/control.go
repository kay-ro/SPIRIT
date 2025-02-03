package gui

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
	"physicsGUI/pkg/minimizer"
	"time"
)

var (
	Problem *minimizer.AsyncMinimiserProblem[float64]
)

func createMinimizerButton() *widget.Button {
	return widget.NewButtonWithIcon("Start Minimizer", theme.MediaPlayIcon(), func() {
		if Problem != nil {
			dialog.ShowError(errors.New("minimizer is already instantiated"), MainWindow)
			return
		}
		errorStr, p := createMinimizerProblem()
		Problem = p

		if Problem != nil {
			closeChan := make(chan struct{}, 1)
			clock := time.Tick(1 * time.Second)

			go minimizeRefreshWorker(Problem, closeChan, clock)

			go func() {
				_ = minimizer.State.Set(2) // minimizer running
				minimizer.FloatMinimizerStagedHC.Minimize(Problem)
				closeChan <- struct{}{}
				_ = minimizer.State.Set(1) // minimizer done, back to ready
				dialog.ShowInformation("Minimizer", "Minimization completed.", MainWindow)
			}()

			log.Println("Minimization started")
			return
		}

		dialogFailedText := fmt.Sprintf("Failed to create minimization problem: %s", errorStr)

		dialog.ShowError(errors.New(dialogFailedText), MainWindow)
	})
}

func createPauseButton() *widget.Button {
	return widget.NewButtonWithIcon("Pause", theme.MediaPauseIcon(), func() {
		if Problem == nil {
			return
		}
		err := Problem.Pause()
		if err != nil {
			dialog.ShowError(err, MainWindow)
		}
	})
}

func createResumeButton() *widget.Button {
	return widget.NewButtonWithIcon("Resume", theme.NavigateNextIcon(), func() {
		if Problem == nil {
			return
		}
		err := Problem.Resume()
		if err != nil {
			dialog.ShowError(err, MainWindow)
		}
	})
}

func createMinimizerStateLabel() *widget.Label {
	label := widget.NewLabel("")

	s, _ := minimizer.State.Get()
	if s == 0 {
		label.SetText("Minimizer Not Ready")
	} else if s == 1 {
		label.SetText("Minimizer Ready")
	}

	minimizer.State.AddListener(binding.NewDataListener(func() {
		state, _ := minimizer.State.Get()
		switch state {
		case 0:
			label.SetText("Minimizer Not Ready")
		case 1:
			label.SetText("Minimizer Ready")
		case 2:
			label.SetText("Minimizer Running")
		case 3:
			label.SetText("Minimizer Paused")
		default:
			label.SetText("Faulty Minimizer State Defined")
		}
	}))

	return label
}
