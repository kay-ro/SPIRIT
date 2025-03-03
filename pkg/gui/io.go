package gui

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"os"
	"physicsGUI/pkg/gui/param"
	"physicsGUI/pkg/io"
	"reflect"
	"strings"
)

func loadFileChooser() {
	// select file
	fileDialog := dialog.NewFileOpen(
		func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, MainWindow)
				return
			}
			if reader == nil {
				return // user aborted
			}
			uri := reader.URI()

			// read file
			data, err := os.ReadFile(uri.Path())
			if err != nil {
				dialog.ShowError(err, MainWindow)
				return
			}

			var config *io.ConfigInformation
			var decodeErr error

			// decode loaded bytes
			if strings.EqualFold(".xml", uri.Extension()) {
				config, decodeErr = io.DecodeXMLFromBytes(data)
			} else if strings.EqualFold(".json", uri.Extension()) {
				config, decodeErr = io.DecodeJSONFromBytes(data)
			} else {
				// if format is not xml or json try binary encoding
				config, decodeErr = io.DecodeGOBFromBytes(data)
			}
			if decodeErr != nil {
				dialog.ShowError(decodeErr, MainWindow)
				return
			}
			for _, value := range config.Parameter {
				if strings.EqualFold(reflect.TypeOf(float64(0)).String(), value.FieldType) {
					fParam := param.GetFloatGroup(value.Group).GetParam(value.Name)
					if val, err := param.StdFloatParser(value.FieldValue); err == nil {
						err = fParam.Set(val)
						if err != nil {
							dialog.ShowError(err, MainWindow)
							return
						}
					} else {
						dialog.ShowError(err, MainWindow)
						return
					}
					fParam.SetCheck(value.UseInFit)
					if value.IsLimited {
						minV, err := param.StdFloatParser(value.FieldMinimum)
						if err != nil {
							dialog.ShowError(err, MainWindow)
						}
						maxV, err := param.StdFloatParser(value.FieldMaximum)
						if err != nil {
							dialog.ShowError(err, MainWindow)
						}
						minP := fParam.GetRelative("min")
						maxP := fParam.GetRelative("max")
						if minP == nil || maxP == nil {
							dialog.ShowInformation("Loading Error", fmt.Sprintf("Mismatching Limitations: Data to load contains limitations for Parameter %s/%s, but current program does not. Limitations will be discarded"), MainWindow)
						}
						err = minP.Set(minV)
						if err != nil {
							dialog.ShowError(err, MainWindow)
							return
						}
						err = maxP.Set(maxV)
						if err != nil {
							dialog.ShowError(err, MainWindow)
							return
						}

					}
				} else if strings.EqualFold(reflect.TypeOf(int(0)).String(), value.FieldType) {
					if val, err := param.StdIntParser(value.FieldValue); err == nil {
						err = param.SetInt(value.Group, value.Name, val)
						if err != nil {
							dialog.ShowError(err, MainWindow)
							return
						}
					} else {
						dialog.ShowError(err, MainWindow)
						return
					}
				} else if strings.EqualFold(reflect.TypeOf(string("")).String(), value.FieldType) {
					if val, err := param.StdStringParser(value.FieldValue); err == nil {
						err = param.SetString(value.Group, value.Name, val)
						if err != nil {
							dialog.ShowError(err, MainWindow)
							return
						}
					} else {
						dialog.ShowError(err, MainWindow)
						return
					}

				} else {
					dialog.ShowError(errors.New(fmt.Sprint(value.FieldType)+" is not supported type"), MainWindow)
					return
				}
			}
		}, MainWindow)

	fileDialog.Show()
}
func saveFileChooser() {
	// setup config

	parameters := make([]io.ParameterInformation, 0)
	// all float parameters
	for _, g := range param.GetFloatKeys() {
		for _, n := range param.GetFloatGroup(g).GetKeys() {
			gParam := param.GetFloatGroup(g).GetParam(n)
			value, gError := gParam.Get()
			if gError != nil {
				dialog.ShowError(gError, MainWindow)
				return
			}
			minP := gParam.GetRelative("min")
			maxP := gParam.GetRelative("max")
			limited := !(minP == nil || maxP == nil)
			minS := ""
			maxS := ""
			if limited {
				minV, err := minP.Get()
				if err != nil {
					dialog.ShowError(err, MainWindow)
					return
				}
				maxV, err := maxP.Get()
				if err != nil {
					dialog.ShowError(err, MainWindow)
					return
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
				dialog.ShowError(gError, MainWindow)
				return
			}
			minP := gParam.GetRelative("min")
			maxP := gParam.GetRelative("max")
			limited := !(minP == nil || maxP == nil)
			minS := ""
			maxS := ""
			if limited {
				minV, err := minP.Get()
				if err != nil {
					dialog.ShowError(err, MainWindow)
					return
				}
				maxV, err := maxP.Get()
				if err != nil {
					dialog.ShowError(err, MainWindow)
					return
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
				dialog.ShowError(gError, MainWindow)
				return
			}
			minP := gParam.GetRelative("min")
			maxP := gParam.GetRelative("max")
			limited := !(minP == nil || maxP == nil)
			minS := ""
			maxS := ""
			if limited {
				minV, err := minP.Get()
				if err != nil {
					dialog.ShowError(err, MainWindow)
					return
				}
				maxV, err := maxP.Get()
				if err != nil {
					dialog.ShowError(err, MainWindow)
					return
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

	config := &io.ConfigInformation{
		Parameter: parameters,
	}

	// select file
	fileDialog := dialog.NewFileSave(
		func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, MainWindow)
				return
			}
			if writer == nil {
				return // user aborted
			}
			var data []byte
			var eError error
			uri := writer.URI()
			if strings.EqualFold(".xml", uri.Extension()) {
				data, eError = io.EncodeXMLToBytes(config)
			} else if strings.EqualFold(".json", uri.Extension()) {
				data, eError = io.EncodeJSONToBytes(config)
			} else {
				data, eError = io.EncodeGOBToBytes(config)
			}
			if eError != nil {
				dialog.ShowError(eError, MainWindow)
				return
			}

			wError := os.WriteFile(uri.Path(), data, 0644)
			if wError != nil {
				dialog.ShowError(err, MainWindow)
				return
			}
		}, MainWindow)
	fileDialog.Show()
}
