package gui

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"physicsGUI/pkg/data"
	"physicsGUI/pkg/function"
	"physicsGUI/pkg/gui/graph"
	"physicsGUI/pkg/gui/helper"
	"physicsGUI/pkg/gui/param"
	"physicsGUI/pkg/minimizer"
	"physicsGUI/pkg/physics"
	"physicsGUI/pkg/trigger"
	"time"
)

var (
	// App reference
	App        fyne.App
	MainWindow fyne.Window

	functionMap       = make(map[string]*function.Function)
	graphMap          = make(map[string]*graph.GraphCanvas)
	floatParameterMap = make(map[string]*param.Parameter[float64])
)

// Start GUI (function is blocking)
func Start() {
	App = app.NewWithID("GUI-Physics")
	App.Settings().SetTheme(theme.DarkTheme()) //TODO WIP to fix invisable while parameter lables
	MainWindow = App.NewWindow("Physics GUI")

	mainWindow()
}

func addDataset(reader io.ReadCloser, uri fyne.URI, err error) function.Points {
	if err != nil {
		dialog.ShowError(err, MainWindow)
		return nil
	}
	if reader == nil {
		return nil // user canceled
	}
	defer func() {
		if err := reader.Close(); err != nil {
			log.Println("error while closing dialog:", err)
		}
	}()

	// read file
	bytes, err := io.ReadAll(reader)
	if err != nil {
		dialog.ShowError(err, MainWindow)
		return nil
	}

	// get filename
	filename := filepath.Base(uri.Name())

	// handle import
	points, err := data.Parse(bytes)
	if err != nil {
		dialog.ShowError(err, MainWindow)
		return nil
	}

	if len(points) == 0 {
		dialog.ShowError(errors.New("no data"), MainWindow)
		return nil
	}

	// show success message
	dialog.ShowInformation("Import successful",
		fmt.Sprintf("File '%s' imported", filename),
		MainWindow)

	return points
}

func createMinimizeButton() *widget.Button {
	btnMinimize := widget.NewButton("Minimize", func() {
		problem := createMinimizerProblem()
		if problem == nil {
			dialog.ShowInformation("Minimizer", "Minimisation Failed: Failed to create minimization Problem",
				MainWindow)
			return
		}
		closeChan := make(chan struct{}, 1)
		clock := time.Tick(1 * time.Second)
		go minimizeRefreshWorker(problem, closeChan, clock)

		go func() {
			minimizer.FloatMinimizerPLLS.Minimize(problem)
			closeChan <- struct{}{}
			dialog.ShowInformation("Minimizer", "Minimisation completed",
				MainWindow)
		}()
	})
	return btnMinimize
}
func minimizeRefreshWorker(problem *minimizer.AsyncMinimiserProblem[float64], close <-chan struct{}, clock <-chan time.Time) {
	for {
		select {
		case <-close:
			return
		case <-clock:
			if parameters, err := problem.GetCurrentParameters(); err != nil {
				continue
			} else {
				_ = floatParameterMap["edenA"].Set(parameters[0])
				_ = floatParameterMap["eden1"].Set(parameters[1])
				_ = floatParameterMap["eden2"].Set(parameters[2])
				_ = floatParameterMap["edenB"].Set(parameters[3])
				_ = floatParameterMap["thickness1"].Set(parameters[4])
				_ = floatParameterMap["thickness2"].Set(parameters[5])
				_ = floatParameterMap["roughnessA1"].Set(parameters[6])
				_ = floatParameterMap["roughness12"].Set(parameters[7])
				_ = floatParameterMap["roughness1B"].Set(parameters[8])
				_ = floatParameterMap["deltaQ"].Set(parameters[9])
				_ = floatParameterMap["background"].Set(parameters[10])
				_ = floatParameterMap["scaling"].Set(parameters[11])
			}

		}
	}
}

func createMinimizerProblem() *minimizer.AsyncMinimiserProblem[float64] {
	eden, err := param.GetFloats("eden")
	if err != nil {
		return nil
	}
	d, err := param.GetFloats("thick")
	if err != nil {
		return nil
	}
	sigma, err := param.GetFloats("rough")
	if err != nil {
		return nil
	}

	// get general parameters
	delta, err := param.GetFloat("general", "deltaq")
	if err != nil {
		return nil
	}
	background, err := param.GetFloat("general", "background")
	if err != nil {
		return nil
	}
	scaling, err := param.GetFloat("general", "scaling")
	if err != nil {
		return nil
	}

	parameters := make([]float64, 0)
	parameters = append(parameters, eden...)
	parameters = append(parameters, d...)
	parameters = append(parameters, sigma...)
	parameters = append(parameters, delta)
	parameters = append(parameters, background)
	parameters = append(parameters, scaling)

	dataTracks := graphMap["sld"].GetDataTracks()
	if len(dataTracks) == 0 {
		return nil
	}
	dataTrack := dataTracks[0]
	// Define error function
	errorFunction := func(params []float64) float64 {
		edenErr := params[0:4]
		dErr := params[4:6]
		sigmaErr := params[6:9]
		deltaErr := params[9]
		backgroundErr := params[10]
		scalingErr := params[11]

		edenPoints, err := physics.GetEdensities(edenErr, dErr, sigmaErr, ZNUMBER)
		if err != nil {
			fmt.Println("Error while calculating edensities:", err)
			return math.MaxFloat64
		}
		sld := make([]float64, len(edenPoints))
		for i, e := range edenPoints {
			sld[i] = e.Y * ELECTRON_RADIUS
		}

		deltaz := 0.0
		if edenPoints != nil && len(edenPoints) > 1 {
			deltaz = edenPoints[1].X - edenPoints[0].X
		}

		// calculate intensity
		modifiedQzAxis := make([]float64, len(qzAxis))
		copy(modifiedQzAxis, qzAxis)
		helper.Map(modifiedQzAxis, func(xPoint float64) float64 { return xPoint + deltaErr })
		intensity := physics.CalculateIntensity(qzAxis, deltaz, sld, &physics.IntensityOptions{
			Background: backgroundErr,
			Scaling:    scalingErr,
		})

		// creates list with intensity points based on edenPoints x and error and calculated intensity as y
		intensityPoints := make(function.Points, QZNUMBER)
		for i := range intensity {
			intensityPoints[i] = &function.Point{
				X:     qzAxis[i],
				Y:     intensity[i],
				Error: 0.0,
			}
			function.Magie(intensityPoints[i])
		}
		intensityFunction := function.NewFunction(intensityPoints, function.INTERPOLATION_LINEAR)
		intensityFunction.Range(dataTrack.Scope.MinX, dataTrack.Scope.MaxX)
		reelD, _ := dataTrack.Model(0, false)
		_, interI := intensityFunction.Model(len(reelD), false)
		functionMap["test"].SetData(interI)
		diff := 0.0
		for i := range reelD {
			diff += math.Abs(reelD[i].Y - interI[i].Y)
		}
		return diff
	}
	minima := make([]float64, len(parameters))
	maxima := make([]float64, len(parameters))
	for i := range minima {
		minima[i] = -math.MaxFloat64
		maxima[i] = math.MaxFloat64
	}
	return minimizer.NewProblem(parameters, minima, maxima, errorFunction, &minimizer.MinimiserConfig{
		LoopCount:     1e2,
		ParallelReads: true,
	})
}

func createImportButton(window fyne.Window) *widget.Button {
	return widget.NewButton("Import Data", func() {
		// open dialog
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			addDataset(reader, reader.URI(), err)
			// TODO: HELP
		}, window)
	})
}

// register functions which can be used for graph plotting
func registerFunctions() {
	functionMap["sld"] = function.NewEmptyFunction(function.INTERPOLATION_NONE)
	functionMap["eden"] = function.NewEmptyFunction(function.INTERPOLATION_NONE)
	functionMap["test"] = function.NewFunction(function.Points{}, function.INTERPOLATION_NONE)
}

// creates the graph containers for the different graphs
func registerGraphs() *fyne.Container {
	graphMap["sld"] = graph.NewGraphCanvas(&graph.GraphConfig{
		Title:     "Intensity Graph",
		IsLog:     false,
		Functions: function.Functions{functionMap["sld"]},
	})

	graphMap["eden"] = graph.NewGraphCanvas(&graph.GraphConfig{
		Title:     "Edensity Graph",
		IsLog:     false,
		Functions: function.Functions{functionMap["eden"]},
	})

	graphMap["test"] = graph.NewGraphCanvas(&graph.GraphConfig{
		Title:     "Test Graph",
		IsLog:     true,
		Functions: function.Functions{functionMap["test"]},
	})

	return container.NewGridWithColumns(2, graphMap["eden"], graphMap["sld"], graphMap["test"])
}

// creates and registers the parameter and adds them to the parameter repository
func registerParams() *fyne.Container {
	edenA, pEdenA := param.FloatMinMax("eden", "Eden a", 0.0)
	eden1, pEden1 := param.FloatMinMax("eden", "Eden 1", 0.346197)
	eden2, pEden2 := param.FloatMinMax("eden", "Eden 2", 0.458849)
	edenB, pEdenB := param.FloatMinMax("eden", "Eden b", 0.334000)

	roughnessA1, pRoughnessA1 := param.FloatMinMax("rough", "Roughness a/1", 3.39544)
	roughness12, pRoughness12 := param.FloatMinMax("rough", "Roughness 1/2", 2.15980)
	roughness2B, pRoughness2B := param.FloatMinMax("rough", "Roughness 2/b", 3.90204)

	thickness1, pThickness1 := param.FloatMinMax("thick", "Thickness 1", 14.2657)
	thickness2, pThickness2 := param.FloatMinMax("thick", "Thickness 2", 10.6906)

	deltaQ, pDeltaQ := param.Float("general", "deltaq", 0.0)
	background, pBackground := param.Float("general", "background", 10e-9)
	scaling, pScaling := param.Float("general", "scaling", 1.0)

	floatParameterMap["edenA"] = pEdenA
	floatParameterMap["eden1"] = pEden1
	floatParameterMap["eden2"] = pEden2
	floatParameterMap["edenB"] = pEdenB
	floatParameterMap["thickness1"] = pThickness1
	floatParameterMap["thickness2"] = pThickness2
	floatParameterMap["roughnessA1"] = pRoughnessA1
	floatParameterMap["roughness12"] = pRoughness12
	floatParameterMap["roughness1B"] = pRoughness2B
	floatParameterMap["deltaQ"] = pDeltaQ
	floatParameterMap["background"] = pBackground
	floatParameterMap["scaling"] = pScaling

	return container.NewVBox(
		container.NewGridWithColumns(4, edenA, eden1, eden2, edenB),
		container.NewGridWithColumns(4, roughnessA1, roughness12, roughness2B),
		container.NewGridWithColumns(4, thickness1, thickness2),
		container.NewGridWithColumns(4, deltaQ, background, scaling),
	)
}

func onDrop(position fyne.Position, uri []fyne.URI) {
	for mapIdentifier, u := range graphMap {
		if u.MouseInCanvas(position) {
			fmt.Println("Dropped on graph:", mapIdentifier)

			for _, v := range uri {
				rc, err := os.OpenFile(v.Path(), os.O_RDONLY, 0666)
				if err != nil {
					dialog.ShowError(err, MainWindow)
					return
				}

				if points := addDataset(rc, v, nil); points != nil {
					points.Magie()
					newFunction := function.NewFunction(points, function.INTERPOLATION_NONE)
					graphMap[mapIdentifier].AddDataTrack(newFunction)
				}
			}
			return
		}
	}
}

// mainWindow builds and renders the main GUI content, it will show and run the main window,
// which is a blocking command [fyne.Window.ShowAndRun]
func mainWindow() {
	registerFunctions()

	content := container.NewBorder(
		container.NewVBox(
			container.NewHBox(
				createMinimizeButton(),
			),
			helper.CreateSeparator(),
		), // top
		nil, // bottom
		nil, // left
		nil, // right

		container.NewVSplit(
			registerGraphs(),
			registerParams(),
		),
	)

	// set onchange function for recalculating data
	trigger.SetOnChange(RecalculateData)

	MainWindow.Resize(fyne.NewSize(1000, 500))
	MainWindow.SetContent(content)
	MainWindow.SetOnDropped(onDrop)

	MainWindow.ShowAndRun()
}

const (
	ELECTRON_RADIUS = 2.81e-5 // classical electron radius in angstrom
	ZNUMBER         = 150
	QZNUMBER        = 500
)

var qzAxis = physics.GetDefaultQZAxis(QZNUMBER)

func testFunc() {
	counter := 11

	d := make(function.Points, counter)

	for i := 0; i < counter; i++ {
		d[i] = &function.Point{
			X:     float64(i),
			Y:     math.Pow(float64(i), 2),
			Error: 1,
		}
	}

	functionMap["sld"].SetData(d)
	functionMap["eden"].SetData(d)
}

// RecalculateData recalculates the data for the sld and eden graphs
func RecalculateData() {
	// Get current parameter groups
	eden, err := param.GetFloats("eden")
	if err != nil {
		return
	}
	d, err := param.GetFloats("thick")
	if err != nil {
		return
	}
	sigma, err := param.GetFloats("rough")
	if err != nil {
		return
	}

	// get general parameters
	delta, err := param.GetFloat("general", "deltaq")
	if err != nil {
		return
	}
	background, err := param.GetFloat("general", "background")
	if err != nil {
		return
	}
	scaling, err := param.GetFloat("general", "scaling")
	if err != nil {
		return
	}

	// calculate edensity
	edenPoints, err := physics.GetEdensities(eden, d, sigma, ZNUMBER)
	if err != nil {
		fmt.Println("Error while calculating edensities:", err)
		return
	} else {
		functionMap["eden"].SetData(edenPoints)
	}

	// transform points into sld floats
	sld := make([]float64, len(edenPoints))
	for i, e := range edenPoints {
		sld[i] = e.Y * ELECTRON_RADIUS
	}

	deltaz := 0.0
	if edenPoints != nil && len(edenPoints) > 1 {
		deltaz = edenPoints[1].X - edenPoints[0].X
	}

	// calculate intensity
	modifiedQzAxis := make([]float64, len(qzAxis))
	copy(modifiedQzAxis, qzAxis)
	helper.Map(modifiedQzAxis, func(xPoint float64) float64 { return xPoint + delta })
	intensity := physics.CalculateIntensity(qzAxis, deltaz, sld, &physics.IntensityOptions{
		Background: background,
		Scaling:    scaling,
	})

	// creates list with intensity points based on edenPoints x and error and calculated intensity as y
	intensityPoints := make(function.Points, QZNUMBER)
	for i := range intensity {
		intensityPoints[i] = &function.Point{
			X:     qzAxis[i],
			Y:     intensity[i],
			Error: 0.0,
		}
		function.Magie(intensityPoints[i])
		//fmt.Println(*intensityPoints[i])

	}

	functionMap["sld"].SetData(intensityPoints)
}
