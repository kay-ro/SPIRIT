package io

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"physicsGUI/pkg/function"
	"reflect"
)

type PointsExport struct {
	Id     string            `json:"id" xml:"id"`
	Points []function.Points `json:"data" xml:"data"`
}

func ExportXMLToFile(pointsToExport []PointsExport) ([]byte, error) {
	return xml.Marshal(pointsToExport)
}

func ExportJSONToFile(pointsToExport []PointsExport) ([]byte, error) {
	return json.Marshal(pointsToExport)
}

func ExportCSVToFile(pointsToExport []PointsExport) ([]byte, error) {
	var byteBuffer = bytes.NewBuffer(nil)
	pointType := reflect.TypeOf(function.Point{})
	pointLen := pointType.NumField()
	totalTrackCount := 0
	maxLength := 0
	for _, point := range pointsToExport {
		totalTrackCount += len(point.Points)
		for _, points := range point.Points {
			maxLength = max(maxLength, len(points))
		}
	}
	w := csv.NewWriter(byteBuffer)

	// write plot names
	titles := make([]string, totalTrackCount*pointLen)
	offset := 0
	for _, p := range pointsToExport {
		titles[offset] = p.Id
		offset += len(p.Points) * pointLen
	}
	err := w.Write(titles)
	if err != nil {
		return nil, err
	}

	// write point field names
	header := make([]string, 0, totalTrackCount*pointLen)
	pointFields := make([]string, 0, pointLen)
	for i := 0; i < pointLen; i++ {
		pointFields = append(pointFields, pointType.Field(i).Name)
	}
	for i := 0; i < totalTrackCount; i++ {
		header = append(header, pointFields...)
	}
	err = w.Write(header)
	if err != nil {
		return nil, err
	}

	// write point field data
	for i := 0; i < maxLength; i++ {
		line := make([]string, 0, totalTrackCount*pointLen)

		for px := 0; px < len(pointsToExport); px++ {
			for pp := 0; pp < len(pointsToExport[px].Points); pp++ {
				for fid := range pointLen {
					if len(pointsToExport[px].Points[pp]) > i {
						field := reflect.ValueOf(pointsToExport[px].Points[pp][i]).Elem().Field(fid)
						if !field.CanConvert(reflect.TypeOf("")) {
							// resolve if float to String
							if field.Kind() == reflect.Float64 || field.Kind() == reflect.Float32 {
								fVal := field.Float()
								line = append(line, fmt.Sprintf("%g", fVal))
							} else {
								// throw if not treated
								return nil, errors.New(fmt.Sprintf("Could not convert point field %s (%s) to %s represenatation => Export failed", pointType.Field(fid).Name, field.Type().Name(), reflect.TypeOf("").Name()))
							}
						} else {
							sVal := field.Convert(reflect.TypeOf(""))
							line = append(line, sVal.String())
						}
					} else {
						line = append(line, "")
					}

				}
			}
		}

		err = w.Write(line)
		if err != nil {
			return nil, err
		}
	}

	w.Flush()
	return byteBuffer.Bytes(), nil
}

func ExportDefaultToFile(pointsToExport []PointsExport) ([]byte, error) {
	return ExportCSVToFile(pointsToExport) // use csv for default export of points
}
