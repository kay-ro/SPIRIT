package gui

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	io2 "io"
	"os"
	"physicsGUI/pkg/io"
	"strings"
)

func loadFileChooser() {
	// select file
	fileDialog := dialog.NewFileOpen(
		fileLoader, MainWindow)

	fileDialog.Show()
}

func fileLoader(reader fyne.URIReadCloser, err error) {
	if err != nil {
		dialog.ShowError(err, MainWindow)
		return
	}
	if reader == nil {
		return // user aborted
	}
	uri := reader.URI()

	// read file
	data, err := io2.ReadAll(reader)
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
	err = LoadConfig(config, false)
	if err != nil {
		if errors.Is(err, differentParameterVersionError) || errors.Is(err, differentPlotVersionError) {
			callBack := func(b bool) {
				if !b {
					return // abort after version mismatch
				} else {
					_ = LoadConfig(config, true) // perform force load and ignore all errors
					return
				}
			}
			dialog.ShowConfirm("Warning: Version Mismatch", "The versions of this program and the one the file was saved have different parameters/plots.\n Do you want to forcefully load anyways? (please safe parameters before forcing a load)", callBack, MainWindow)
		} else {
			dialog.ShowError(err, MainWindow)
		}
	}

}

func saveFileChooser() {
	// select file
	fileDialog := dialog.NewFileSave(fileSaver, MainWindow)
	fileDialog.Show()
}

func fileSaver(writer fyne.URIWriteCloser, err error) {
	if err != nil {
		dialog.ShowError(err, MainWindow)
		return
	}
	if writer == nil {
		return // user aborted
	}

	// setup config
	config, err := CreateConfig()
	if err != nil {
		dialog.ShowError(err, MainWindow)
		return
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

	wError := os.WriteFile(uri.Path(), data, 0644) // TODO fix magic number eventually
	if wError != nil {
		dialog.ShowError(err, MainWindow)
		return
	}
}
