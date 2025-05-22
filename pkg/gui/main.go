package gui

import (
	"errors"
	"fmt"
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

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
)

var (
	// App reference
	App        fyne.App
	MainWindow fyne.Window

	functionMap = make(map[string]*function.Function)
	graphMap    = make(map[string]*graph.GraphCanvas)
)

// adaption should not be necessary here
// Start GUI (function is blocking)
func Start() {
	App = app.NewWithID("GUI-Physics")
	App.Settings().SetTheme(theme.DarkTheme())
	MainWindow = App.NewWindow("Physics GUI")

	mainWindow()
}

// adaption should not be necessary
// onDrop is called when a file is dropped into the window
// imports the data if a file is dropped on a graph canvas
func onDrop(position fyne.Position, uri []fyne.URI) {
	for mapIdentifier, u := range graphMap {
		if u.MouseInCanvas(position) {
			for _, v := range uri {
				rc, err := os.OpenFile(v.Path(), os.O_RDONLY, 0666)
				if err != nil {
					dialog.ShowError(err, MainWindow)
					return
				}

				if points := addDataset(rc, v, nil); points != nil {
					newFunction := function.NewFunction(points)
					graphMap[mapIdentifier].AddDataTrack(newFunction)
					physics.AlterQZAxis(graphMap[mapIdentifier].GetDataTracks(), mapIdentifier)
				}
			}
			return
		}
	}
}

// adaption should not be necessary here
// parses a given file into a dataset
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

// adaption should not be necessary here
func createFileMenu() *fyne.Menu {
	mnLoad := fyne.NewMenuItem("Load", loadFileChooser)
	mnSave := fyne.NewMenuItem("Save", saveFileChooser)
	mnExport := fyne.NewMenuItem("Export", exportFileChooser)
	return fyne.NewMenu("File", mnLoad, mnSave, mnExport)
}

// adaption should not be necessary here
// mainWindow builds and renders the main GUI content, it will show and run the main window
func mainWindow() {
	registerFunctions()

	content := container.NewBorder(
		container.NewVBox(
			container.NewHBox(
				NewMinimizerControlPanel().Widget(),
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

	MainWindow.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("Program"),
		createFileMenu(),
	))
	MainWindow.Resize(fyne.NewSize(1000, 500))
	MainWindow.SetContent(content)
	MainWindow.SetOnDropped(onDrop)

	MainWindow.ShowAndRun()
}

//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! adapt everything from here !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

// this is also the place where you need to pass:
// all current parameters and all experimental data tracks
func (controlPanel *MinimizerControlPanel) minimizerProblemSetup() error {
	// get parameters + experimental data and put them into minimize()
	edens := param.GetFloatGroup("eden")

	e1 := edens.GetParam("Eden Au")
	e2 := edens.GetParam("Eden Org")
	e3 := edens.GetParam("Eden Bulk")

	// get roughness parameters
	other := param.GetFloatGroup("other")

	roughness := other.GetParam("Roughness")
	coverage := other.GetParam("Coverage")

	// get thickness parameters
	thickness := param.GetFloatGroup("size")

	s1 := thickness.GetParam("Radius")
	s2 := thickness.GetParam("Shell Thickness")
	s3 := thickness.GetParam("z Offset")
	s4 := thickness.GetParam("z Offset Au Org")

	// get general parameters
	general := param.GetFloatGroup("general")

	delta := general.GetParam("deltaq")
	background := general.GetParam("background")
	scaling := general.GetParam("scaling")
	resolution := general.GetParam("resolution")

	if err := controlPanel.minimize(e1, e2, e3, s1, s2, s3, s4, roughness, coverage, delta, background, scaling, resolution); err != nil {
		fmt.Println("Error while minimizing:", err)
		return err
	}
	return nil
}

// the penalty function defines the error we minimize with minuit
// !the order of the parameters needs to fit
func penaltyFunction(fcn *minimizer.MinuitFunction, params []float64) float64 {
	paramCount := 13
	if len(params) != paramCount {
		dialog.ShowError(fmt.Errorf("penaltyFunction has %d parameters but expects %d", len(params), paramCount), MainWindow)
		return math.MaxFloat64
	}

	eden := params[0:3]
	size := params[3:7]
	roughness := params[7]
	coverage := params[8]
	deltaq := params[9]
	background := params[10]
	scaling := params[11]
	resolution := params[12]

	log.Println("params", params)

	edenPoints, err := physics.GetEdensities(eden, size, roughness, coverage)
	if err != nil {
		fmt.Println("Error while calculating edensities:", err)
		return math.MaxFloat64
	}

	intensityPoints := physics.CalculateIntensityPoints(edenPoints, deltaq, &physics.IntensityOptions{
		Background: background,
		Scaling:    scaling,
		Resolution: resolution,
	})

	experimentalData := graphMap["intensity"].GetDataTracks()
	dataTracks := make([]function.Points, len(experimentalData))
	for i, dataTrack := range experimentalData {
		dataTracks[i] = dataTrack.GetData()
	}

	//penalty calculation
	diff, err := physics.Sim2SigRMS(dataTracks, intensityPoints)
	if err != nil {
		dialog.ShowError(err, MainWindow)
	}

	return diff
}

// register functions which can be used for graph plotting
// this is the place to add function plots which are shown in graphs
func registerFunctions() {
	//a function needs to be added to the functionMap using a unique identifier so we can further handle it
	//interpolation mode can be ignored
	functionMap["intensity"] = function.NewEmptyFunction()
	functionMap["eden"] = function.NewEmptyFunction()
}

// creates the graph containers for the different graphs
// this is the place to add graphs to the GUI
func registerGraphs() *fyne.Container {
	graphMap["intensity"] = graph.NewGraphCanvas(&graph.GraphConfig{
		Title:     "Intensity Graph",
		IsLog:     true,
		AdaptDraw: true,
		Functions: function.Functions{functionMap["intensity"]},
		DisplayRange: &graph.GraphRange{
			Min: 0.01,
			Max: math.MaxFloat64,
		},
	})

	graphMap["eden"] = graph.NewGraphCanvas(&graph.GraphConfig{
		Title:     "Edensity Graph",
		IsLog:     false,
		AdaptDraw: false,
		Functions: function.Functions{functionMap["eden"]},
	})

	return container.NewGridWithColumns(2, graphMap["eden"], graphMap["intensity"])
}

// creates and registers the parameter and adds them to the parameter repository
// this is the place to alter parameters:
func registerParams() *fyne.Container {
	eden_au, _ := param.FloatMinMax("eden", "Eden Au", 4.66)
	eden_org, _ := param.FloatMinMax("eden", "Eden Org", 0.45)
	eden_b, _ := param.FloatMinMax("eden", "Eden Bulk", 0.334000)

	roughness, _ := param.FloatMinMax("other", "Roughness", 3.5)
	coverage, _ := param.FloatMinMax("other", "Coverage", 0.3)

	radius, _ := param.FloatMinMax("size", "Radius", 15.0)
	d_shell, _ := param.FloatMinMax("size", "Shell Thickness", 17.0)
	z_offset, _ := param.FloatMinMax("size", "z Offset", 0.0)
	z_offset_au_org, _ := param.FloatMinMax("size", "z Offset Au Org", 0.0)

	deltaQ, _ := param.FloatMinMax("general", "deltaq", 0.0)
	background, _ := param.FloatMinMax("general", "background", 1.0e-8)
	scaling, _ := param.FloatMinMax("general", "scaling", 1.0)
	resolution, _ := param.FloatMinMax("general", "resolution", 100.0)

	containers := container.NewVBox(
		container.NewGridWithColumns(4, eden_au, eden_org, eden_b),
		container.NewGridWithColumns(4, roughness, coverage),
		container.NewGridWithColumns(4, radius, d_shell, z_offset, z_offset_au_org),
		container.NewGridWithColumns(4, deltaQ, background, scaling, resolution),
	)

	//makes a scrollbar for the parameters
	con2 := container.NewScroll(containers)
	con2.SetMinSize(fyne.NewSize(300, 300))
	return container.NewStack(con2)
}

// Insert your adapted physical calculations and parameters here!
// RecalculateData recalculates the data for the current graphs
// current parameter values need to be fetched, the physical calculations done and resulting points set to the functions
func RecalculateData() {
	// Get current parameter groups
	eden, err := param.GetFloats("eden")
	if err != nil {
		log.Println("Error while getting eden parameters:", err)
		return
	}
	size, err := param.GetFloats("size")
	if err != nil {
		log.Println("Error while getting thickness parameters:", err)
		return
	}
	roughness, err := param.GetFloat("other", "Roughness")
	if err != nil {
		log.Println("Error while getting roughness parameters:", err)
		return
	}
	coverage, err := param.GetFloat("other", "Coverage")
	if err != nil {
		log.Println("Error while getting roughness parameters:", err)
		return
	}

	// get general parameters
	deltaq, err := param.GetFloat("general", "deltaq")
	if err != nil {
		log.Println("Error while getting deltaq parameter:", err)
		return
	}
	background, err := param.GetFloat("general", "background")
	if err != nil {
		log.Println("Error while getting background parameter:", err)
		return
	}
	scaling, err := param.GetFloat("general", "scaling")
	if err != nil {
		log.Println("Error while getting scaling parameter:", err)
		return
	}
	resolution, err := param.GetFloat("general", "resolution")
	if err != nil {
		log.Println("Error while getting resolution parameter:", err)
		return
	}
	// calculate edensity
	edenPoints, err := physics.GetEdensities(eden, size, roughness, coverage)
	if err != nil {
		log.Println("Error while calculating edensities:", err)
		return
	} else {
		functionMap["eden"].SetData(edenPoints)
	}

	intensityPoints := physics.CalculateIntensityPoints(edenPoints, deltaq, &physics.IntensityOptions{
		Background: background,
		Scaling:    scaling,
		Resolution: resolution,
	})

	functionMap["intensity"].SetData(intensityPoints)
}
