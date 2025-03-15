package gui

import (
	"crypto/md5"
	"errors"
	"fmt"
	"maps"
	"physicsGUI/pkg/function"
	"physicsGUI/pkg/gui/param"
	"physicsGUI/pkg/io"
	"reflect"
	"slices"
	"sort"
	"strings"
)

var differentParameterVersionError = errors.New("different parameter version")
var differentPlotVersionError = errors.New("different plot version")

func makeVersionCheckSum(blocks []string) []byte {
	hasher := md5.New()
	hasher.Write([]byte(strings.Join(blocks, string(byte(0)))))
	return hasher.Sum(nil)
}

func getProgramParameterKeys() []string {
	// get all group names
	groupFloatKeys := param.GetFloatKeys()
	groupIntKeys := param.GetIntKeys()
	groupStringKeys := param.GetStringKeys()
	// Sort names to prevent change in version number caused by different adding order
	sort.Strings(groupFloatKeys)
	sort.Strings(groupIntKeys)
	sort.Strings(groupStringKeys)

	var keys []string = nil
	// Add float Parameters
	for i := range groupFloatKeys {
		fParamKeys := param.GetFloatGroup(groupFloatKeys[i]).GetKeys()
		// Sort names to prevent change in version number caused by different adding order
		sort.Strings(fParamKeys)
		for j := range fParamKeys {
			keys = append(keys, groupFloatKeys[i]+string(byte(255))+fParamKeys[j])
		}
	}
	// Add int Parameters
	for i := range groupIntKeys {
		iParamKeys := param.GetIntGroup(groupIntKeys[i]).GetKeys()
		// Sort names to prevent change in version number caused by different adding order
		sort.Strings(iParamKeys)
		for j := range iParamKeys {
			keys = append(keys, groupIntKeys[i]+string(byte(255))+iParamKeys[j])
		}
	}

	// Add string Parameters
	for i := range groupStringKeys {
		sParamKeys := param.GetStringGroup(groupStringKeys[i]).GetKeys()
		// Sort names to prevent change in version number caused by different adding order
		sort.Strings(sParamKeys)
		for j := range sParamKeys {
			keys = append(keys, groupStringKeys[i]+string(byte(255))+sParamKeys[j])
		}
	}
	return keys
}

func getProgramPlotKeys() []string {
	keys := slices.Collect(maps.Keys(graphMap))
	sort.Strings(keys)
	return keys
}

func LoadConfig(config *io.ConfigInformation, forceLoad bool) error {
	// check parameter version indicator skipped in force Load
	if !forceLoad && !slices.Equal(makeVersionCheckSum(getProgramParameterKeys()), config.ParameterVersionIndicator) {
		return differentParameterVersionError
	}

	// load Parameter information
	err := loadParameterInformation(config.Parameter)
	if err != nil {
		return err
	}

	// check plot version indicator skipped in force Load
	if !forceLoad && !slices.Equal(makeVersionCheckSum(getProgramPlotKeys()), config.PlotVersionIndicator) {
		return differentPlotVersionError
	}

	// load Plot information
	err = loadPlotInformation(config.Plot)
	if err != nil {
		return err
	}

	return nil
}

func loadParameterInformation(paramInfo []io.ParameterInformation) error {
	// load parameters
	for _, value := range paramInfo {
		if strings.EqualFold(reflect.TypeOf(float64(0)).String(), value.FieldType) {
			fpGroup := param.GetFloatGroup(value.Group)
			if fpGroup == nil {
				fmt.Printf("Could not load %s no such parameter group (float64) in program -> Skipped", value.Group)
				continue
			}
			fParam := fpGroup.GetParam(value.Name)
			if fParam == nil {
				fmt.Printf("Could not load %s no such parameter(float64) in program -> Skipped", value.Group+"/"+value.Name)
				continue
			}
			if val, err := param.StdFloatParser(value.FieldValue); err == nil {
				err = fParam.Set(val)
				if err != nil {
					return err
				}
			} else {
				return err
			}
			fParam.SetCheck(value.UseInFit)
			if value.IsLimited {
				minV, err := param.StdFloatParser(value.FieldMinimum)
				if err != nil {
					return err
				}
				maxV, err := param.StdFloatParser(value.FieldMaximum)
				if err != nil {
					return err
				}
				minP := fParam.GetRelative("min")
				maxP := fParam.GetRelative("max")
				print(fParam.Widget().Text)
				if minP == nil || maxP == nil {
					fmt.Printf("Could not set min/max for parameter(float64) %s in program, no such fields -> Skipped", value.Name)
					continue
				} else {
					err = minP.Set(minV)
					if err != nil {
						return err
					}
					err = maxP.Set(maxV)
					if err != nil {
						return err
					}
				}

			}
		} else if strings.EqualFold(reflect.TypeOf(int(0)).String(), value.FieldType) {
			if val, err := param.StdIntParser(value.FieldValue); err == nil {
				err = param.SetInt(value.Group, value.Name, val)
				if err != nil {
					if errors.Is(err, param.ErrParameterNotFound) {
						fmt.Printf("Could not load %s no such parameter(int) in program -> Skipped", value.Group+"/"+value.Name)
						continue
					} else {
						return err
					}
				}
			} else {
				return err
			}
		} else if strings.EqualFold(reflect.TypeOf(string("")).String(), value.FieldType) {
			if val, err := param.StdStringParser(value.FieldValue); err == nil {
				err = param.SetString(value.Group, value.Name, val)
				if err != nil {
					if errors.Is(err, param.ErrParameterNotFound) {
						fmt.Printf("Could not load %s no such parameter(string) in program -> Skipped", value.Group+"/"+value.Name)
						continue
					} else {
						return err
					}
				}
			} else {
				return err
			}

		} else {
			fmt.Printf("Type: %s is not supported -> Skipped", value.FieldType)
		}
	}
	return nil
}

func loadPlotInformation(paramInfo []io.PlotInformation) error {
	for _, information := range paramInfo {
		if _, ok := graphMap[information.Name]; !ok {
			fmt.Printf("Could not load %s no such plot in program", information.Name)
			continue
		}
		for i := 0; i < len(information.DataTracks); i++ {
			fcn := function.NewFunction(information.DataTracks[i].Points)
			scopeCopy := information.DataTracks[i].Scope
			fcn.Scope = &scopeCopy
			graphMap[information.Name].AddDataTrack(fcn)
		}
	}
	return nil
}

func CreateConfig() (*io.ConfigInformation, error) {

	// create ParameterInformation
	parameters, err := createParameterInformation()
	if err != nil {
		return nil, err
	}

	// create PlotInformation
	plot, err := createPlotInformation()
	if err != nil {
		return nil, err
	}

	return &io.ConfigInformation{
		PlotVersionIndicator:      makeVersionCheckSum(getProgramPlotKeys()),
		Plot:                      plot,
		ParameterVersionIndicator: makeVersionCheckSum(getProgramParameterKeys()),
		Parameter:                 parameters,
	}, nil
}

func createParameterInformation() ([]io.ParameterInformation, error) {
	parameters := make([]io.ParameterInformation, 0)
	// all float parameters
	for _, g := range param.GetFloatKeys() {
		for _, n := range param.GetFloatGroup(g).GetKeys() {
			gParam := param.GetFloatGroup(g).GetParam(n)
			value, gError := gParam.Get()
			if gError != nil {
				return nil, gError
			}
			minP := gParam.GetRelative("min")
			maxP := gParam.GetRelative("max")
			limited := !(minP == nil || maxP == nil)
			minS := ""
			maxS := ""
			if limited {
				minV, err := minP.Get()
				if err != nil {
					return nil, err
				}
				maxV, err := maxP.Get()
				if err != nil {
					return nil, err
				}
				minS = param.StdFloatFormater(minV)
				maxS = param.StdFloatFormater(maxV)
			}

			parameters = append(parameters, io.ParameterInformation{
				Group:        g,
				Name:         n,
				FieldValue:   param.StdFloatFormater(value),
				FieldType:    reflect.TypeOf(value).String(),
				UseInFit:     gParam.IsChecked(),
				IsLimited:    limited,
				FieldMinimum: minS,
				FieldMaximum: maxS,
			})
		}
	}

	for _, g := range param.GetStringKeys() {
		for _, n := range param.GetStringGroup(g).GetKeys() {
			gParam := param.GetStringGroup(g).GetParam(n)
			value, gError := gParam.Get()
			if gError != nil {
				return nil, gError
			}
			minP := gParam.GetRelative("min")
			maxP := gParam.GetRelative("max")
			limited := !(minP == nil || maxP == nil)
			minS := ""
			maxS := ""
			if limited {
				minV, err := minP.Get()
				if err != nil {
					return nil, err
				}
				maxV, err := maxP.Get()
				if err != nil {
					return nil, err
				}
				minS = param.StdStringFormater(minV)
				maxS = param.StdStringFormater(maxV)
			}

			parameters = append(parameters, io.ParameterInformation{
				Group:        g,
				Name:         n,
				FieldValue:   param.StdStringFormater(value),
				FieldType:    reflect.TypeOf(value).String(),
				UseInFit:     gParam.IsChecked(),
				IsLimited:    limited,
				FieldMinimum: minS,
				FieldMaximum: maxS,
			})
		}
	}

	for _, g := range param.GetIntKeys() {
		for _, n := range param.GetIntGroup(g).GetKeys() {
			gParam := param.GetIntGroup(g).GetParam(n)
			value, gError := gParam.Get()
			if gError != nil {
				return nil, gError
			}
			minP := gParam.GetRelative("min")
			maxP := gParam.GetRelative("max")
			limited := !(minP == nil || maxP == nil)
			minS := ""
			maxS := ""
			if limited {
				minV, err := minP.Get()
				if err != nil {
					return nil, err
				}
				maxV, err := maxP.Get()
				if err != nil {
					return nil, err
				}
				minS = param.StdIntFormater(minV)
				maxS = param.StdIntFormater(maxV)
			}

			parameters = append(parameters, io.ParameterInformation{
				Group:        g,
				Name:         n,
				FieldValue:   param.StdIntFormater(value),
				FieldType:    reflect.TypeOf(value).String(),
				UseInFit:     gParam.IsChecked(),
				IsLimited:    limited,
				FieldMinimum: minS,
				FieldMaximum: maxS,
			})
		}
	}

	return parameters, nil
}

func createPlotInformation() ([]io.PlotInformation, error) {
	plotInfos := make([]io.PlotInformation, 0, len(graphMap))

	for key, plot := range graphMap {
		dataTracks := plot.GetDataTracks()
		funcInfos := make([]io.FunctionInformation, 0, len(dataTracks))
		for i := 0; i < len(dataTracks); i++ {
			scopeCopy := *dataTracks[i].Scope // this should copy the struct

			funcInfo := io.FunctionInformation{
				Points: dataTracks[i].GetData(),
				Scope:  scopeCopy,
			}
			funcInfos = append(funcInfos, funcInfo)
		}
		plotInfo := io.PlotInformation{
			Name:       key,
			DataTracks: funcInfos,
		}
		plotInfos = append(plotInfos, plotInfo)
	}

	return plotInfos, nil
}
