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
	"slices"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	minuit "github.com/empack/minuit2go/pkg"
)

var (
	// App reference
	App        fyne.App
	MainWindow fyne.Window

	functionMap = make(map[string]*function.Function)
	graphMap    = make(map[string]*graph.GraphCanvas)

	// TODO move to other location
	errorFunction func(parameter []float64) float64
	groupSequence []string
	dataTracks    = make(map[string]function.Functions)
)

// Start GUI (function is blocking)
func Start() {
	App = app.NewWithID("GUI-Physics")
	App.Settings().SetTheme(theme.DarkTheme()) //TODO WIP to fix invisable while parameter lables
	MainWindow = App.NewWindow("Physics GUI")

	mainWindow()
}

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

func createMinimizeButton() *widget.Button {
	return widget.NewButton("Minimize", func() {
		if errorFunction == nil {
			dialog.ShowInformation("Minimizer Setup", "No error function defined", MainWindow)
			return
		}
		problemFunction := minimizer.NewDummyFCN(errorFunction)
		groupSizes := []int{}

		// load DataTracks
		for n := range graphMap {
			dataTracks[n] = graphMap[n].GetDataTracks()
		}

		// Setup parameters
		userParameters := minuit.NewEmptyMnUserParameters()
		for _, group := range groupSequence {
			vals, err := param.GetFloats(group)
			if err != nil {
				dialog.ShowInformation("Minimizer Setup", fmt.Sprintf("Failed to get value floats of group \"%s\" for minimisation setup", group), MainWindow)
				return
			}
			mins, err := param.GetFloatMinima(group)
			if err != nil {
				dialog.ShowInformation("Minimizer Setup", fmt.Sprintf("Failed to get minima floats of group \"%s\" for minimisation setup", group), MainWindow)
				return
			}
			maxs, err := param.GetFloatMaximas(group)
			if err != nil {
				dialog.ShowInformation("Minimizer Setup", fmt.Sprintf("Failed to get maxima floats of group \"%s\" for minimisation setup", group), MainWindow)
				return
			}
			groupSizes = append(groupSizes, len(vals))
			for i := range vals {
				if vals[i] > maxs[i] || vals[i] < mins[i] { // TODO replace with better solution
					userParameters.AddFree(fmt.Sprintf("%s-%d", group, i), vals[i], 0)
					continue
				}

				if vals[i] == mins[i] && vals[i] == maxs[i] {
					userParameters.Add(fmt.Sprintf("%s-%d", group, i), vals[i])
				} else if mins[i] == -math.MaxFloat64 && maxs[i] == math.MaxFloat64 {
					//TODO error?
					userParameters.AddFree(fmt.Sprintf("%s-%d", group, i), vals[i], 0)
				} else {
					//TODO error?
					userParameters.AddLimited(fmt.Sprintf("%s-%d", group, i), vals[i], 0, mins[i], maxs[i])
				}
			}
		}

		mingrad := minuit.NewMnMigradWithParameters(problemFunction, userParameters)
		minimum, err := mingrad.Minimize()
		if err != nil {
			dialog.ShowError(err, MainWindow)
			return
		}

		var groupID int = 0
		var offset int = 0
		var collection []float64 = make([]float64, groupSizes[0])
		for i, minimizedParameter := range minimum.UserParameters().Parameters() {
			if i >= groupSizes[groupID]+offset {
				err := param.SetFloats(groupSequence[groupID], collection)
				if err != nil {
					dialog.ShowInformation("Minimize", "Failed to write minimized Data to screen", MainWindow)
					return
				}
				groupID++
				offset = i
				collection = make([]float64, groupSizes[groupID])
			}
			collection[i-offset] = minimizedParameter.Value()
		}
		//TODO check if last collection is Set

		dialog.ShowInformation("Minimization Completed", fmt.Sprintf("Minimization Stats:\n Error function calls: %f \n Remaining error: %f", minimum.Nfcn(), minimum.Fval()), MainWindow)
	})
}

func minimizeRefreshWorker(problem *minimizer.AsyncMinimiserProblem[float64], close <-chan struct{}, clock <-chan time.Time) {
	for {
		select {
		case <-close:
			return
		case <-clock:
			parameters, err := problem.GetCurrentParameters()
			if err != nil {
				continue
			}

			param.SetFloats("eden", []float64{parameters[0], parameters[1], parameters[2], parameters[3]})
			param.SetFloats("thick", []float64{parameters[4], parameters[5]})
			param.SetFloats("rough", []float64{parameters[6], parameters[7], parameters[8]})

			param.SetFloat("general", "deltaq", parameters[9])
			param.SetFloat("general", "background", parameters[10])
			param.SetFloat("general", "scaling", parameters[11])
		}
	}
}

func createMinimizerProblem() (string, *minimizer.AsyncMinimiserProblem[float64]) {
	eden, err := param.GetFloats("eden")
	if err != nil {
		return "EDEN FLOATS MISS", nil
	}
	edenMinima, err := param.GetFloatMinima("eden")
	if err != nil {
		return "EDEN MINIMA MISS", nil
	}
	edenMaxima, err := param.GetFloatMaximas("eden")
	if err != nil {
		return "EDEN MAXIMA MISS", nil
	}

	d, err := param.GetFloats("thick")
	if err != nil {
		return "THICK FLOATS MISS", nil
	}
	dMinima, err := param.GetFloatMinima("thick")
	if err != nil {
		return "THICK MINIMA MISS", nil
	}
	dMaxima, err := param.GetFloatMaximas("thick")
	if err != nil {
		return "THICK MAXIMA MISS", nil
	}

	sigma, err := param.GetFloats("rough")
	if err != nil {
		return "ROUGH FLOATS MISS", nil
	}
	sigmaMinima, err := param.GetFloatMinima("rough")
	if err != nil {
		return "ROUGH MINIMA MISS", nil
	}
	sigmaMaxima, err := param.GetFloatMaximas("rough")
	if err != nil {
		return "ROUGH MAXIMA MISS", nil
	}

	// get general parameters
	delta, err := param.GetFloat("general", "deltaq")
	if err != nil {
		return "(GENERAL) DELTAQ MISS", nil
	}
	deltaMinima := -math.MaxFloat64
	deltaMaxima := math.MaxFloat64

	background, err := param.GetFloat("general", "background")
	if err != nil {
		return "(GENERAL) BACKGROUND MISS", nil
	}
	backgroundMinima := 0.0
	backgroundMaxima := math.MaxFloat64

	scaling, err := param.GetFloat("general", "scaling")
	if err != nil {
		return "(GENERAL) SCALING MISS", nil
	}
	scalingMinima := 0.0
	scalingMaxima := math.MaxFloat64

	parameters := slices.Concat(eden, d, sigma)
	parameters = append(parameters, delta, background, scaling)
	minimas := slices.Concat(edenMinima, dMinima, sigmaMinima)
	minimas = append(minimas, deltaMinima, backgroundMinima, scalingMinima)
	maximas := slices.Concat(edenMaxima, dMaxima, sigmaMaxima)
	maximas = append(maximas, deltaMaxima, backgroundMaxima, scalingMaxima)

	dataTracks := graphMap["intensity"].GetDataTracks()
	if len(dataTracks) == 0 {
		return "NO DATA POINTS IN GRAPH", nil
	}
	dataTrack := dataTracks[0]

	// Define error function
	penaltyFunction := func(params []float64) float64 {
		edenErr := params[0:4]
		dErr := params[4:6]
		sigmaErr := params[6:9]
		deltaErr := params[9]
		backgroundErr := params[10]
		scalingErr := params[11]

		edenPoints, err := physics.GetEdensities(edenErr, dErr, sigmaErr)
		if err != nil {
			fmt.Println("Error while calculating edensities:", err)
			return math.MaxFloat64
		}

		intensityPoints := physics.CalculateIntensityPoints(edenPoints, deltaErr, &physics.IntensityOptions{
			Background: backgroundErr,
			Scaling:    scalingErr,
		})

		intensityFunction := function.NewFunction(intensityPoints, function.INTERPOLATION_LINEAR)

		dataModel := dataTrack.Model(0, false)

		diff := 0.0
		for i := range dataModel {
			iy, err := intensityFunction.Eval(dataModel[i].X)
			if err != nil {
				fmt.Println("Error while calculating intensity:", err)
			}
			diff += math.Pow(dataModel[i].Y-iy, 2)
		}
		return diff
	}

	return "", minimizer.NewProblem(parameters, minimas, maximas, penaltyFunction, &minimizer.MinimiserConfig{
		LoopCount:     1e6,
		ParallelReads: true,
	})
}

// register functions which can be used for graph plotting
func registerFunctions() {
	functionMap["intensity"] = function.NewEmptyFunction(function.INTERPOLATION_NONE)
	functionMap["eden"] = function.NewEmptyFunction(function.INTERPOLATION_NONE)
	functionMap["test"] = function.NewFunction(function.Points{}, function.INTERPOLATION_NONE)
}

// creates the graph containers for the different graphs
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

	graphMap["test"] = graph.NewGraphCanvas(&graph.GraphConfig{
		Title:     "Test Graph",
		IsLog:     true,
		AdaptDraw: true,
		Functions: function.Functions{functionMap["test"]},
	})

	return container.NewGridWithColumns(2, graphMap["eden"], graphMap["intensity"], graphMap["test"])
}

// creates and registers the parameter and adds them to the parameter repository
func registerParams() *fyne.Container {
	edenA, _ := param.FloatMinMax("eden", "Eden a", 0.0)
	eden1, _ := param.FloatMinMax("eden", "Eden 1", 0.346197)
	eden2, _ := param.FloatMinMax("eden", "Eden 2", 0.458849)
	edenB, _ := param.FloatMinMax("eden", "Eden b", 0.334000)

	roughnessA1, _ := param.FloatMinMax("rough", "Roughness a/1", 3.39544)
	roughness12, _ := param.FloatMinMax("rough", "Roughness 1/2", 2.15980)
	roughness2B, _ := param.FloatMinMax("rough", "Roughness 2/b", 3.90204)

	thickness1, _ := param.FloatMinMax("thick", "Thickness 1", 14.2657)
	thickness2, _ := param.FloatMinMax("thick", "Thickness 2", 10.6906)

	deltaQ, _ := param.Float("general", "deltaq", -0.000305927)
	background, _ := param.Float("general", "background", 1.43793e-7)
	scaling, _ := param.Float("general", "scaling", 0.888730)

	return container.NewVBox(
		container.NewGridWithColumns(4, edenA, eden1, eden2, edenB),
		container.NewGridWithColumns(4, roughnessA1, roughness12, roughness2B),
		container.NewGridWithColumns(4, thickness1, thickness2),
		container.NewGridWithColumns(4, deltaQ, background, scaling),
	)
}

// onDrop is called when a file is dropped into the window and triggers
// an import of the data if the file is dropped on a graph canvas
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

	// Define Error Function
	MinimizerSetup()

	// set onchange function for recalculating data
	trigger.SetOnChange(RecalculateData)

	MainWindow.Resize(fyne.NewSize(1000, 500))
	MainWindow.SetContent(content)
	MainWindow.SetOnDropped(onDrop)

	MainWindow.ShowAndRun()
}

// this test func will create a basic x^2 dataset for testing
// and set it to the intensity and eden graphs
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

	functionMap["intensity"].SetData(d)
	functionMap["eden"].SetData(d)
}

// Setup Minimization
func MinimizerSetup() {
	// Define Sequence of parameters
	groupSequence = []string{"eden", "thick", "rough", "general"}
	errorFunction = func(parameter []float64) float64 {
		eden := parameter[0:4]
		d := parameter[4:6]
		sigma := parameter[6:9]
		delta := parameter[9]
		background := parameter[10]
		scaling := parameter[11]

		edenPoints, err := physics.GetEdensities(eden, d, sigma)
		if err != nil {
			fmt.Println("Error while calculating edensities:", err)
			return math.MaxFloat64
		}

		intensityPoints := physics.CalculateIntensityPoints(edenPoints, delta, &physics.IntensityOptions{
			Background: background,
			Scaling:    scaling,
		})

		intensityFunction := function.NewFunction(intensityPoints, function.INTERPOLATION_LINEAR)

		//
		plotDataTracks, ok := dataTracks["intensity"]
		if !ok {
			fmt.Println("Could not load data Track from plot please check error function or load data in plot:")
			return math.MaxFloat64
		}
		dataModel := plotDataTracks[0].Model(0, false)

		diff := 0.0
		for i := range dataModel {
			iy, err := intensityFunction.Eval(dataModel[i].X)
			if err != nil {
				fmt.Println("Error while calculating intensity:", err)
			}
			diff += math.Pow(dataModel[i].Y-iy, 2)
		}
		return diff
	}
}

// RecalculateData recalculates the data for the intensity and eden graphs
func RecalculateData() {
	// Get current parameter groups
	eden, err := param.GetFloats("eden")
	if err != nil {
		log.Println("Error while getting eden parameters:", err)
		return
	}
	d, err := param.GetFloats("thick")
	if err != nil {
		log.Println("Error while getting thickness parameters:", err)
		return
	}
	sigma, err := param.GetFloats("rough")
	if err != nil {
		log.Println("Error while getting roughness parameters:", err)
		return
	}

	// get general parameters
	delta, err := param.GetFloat("general", "deltaq")
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

	// calculate edensity
	edenPoints, err := physics.GetEdensities(eden, d, sigma)
	if err != nil {
		log.Println("Error while calculating edensities:", err)
		return
	} else {
		functionMap["eden"].SetData(edenPoints)
	}

	intensityPoints := physics.CalculateIntensityPoints(edenPoints, delta, &physics.IntensityOptions{
		Background: background,
		Scaling:    scaling,
	})

	functionMap["intensity"].SetData(intensityPoints)
}
