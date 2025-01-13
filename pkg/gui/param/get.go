package param

// Get single string value based on group and label
func GetString(group, label string) (string, error) {
	if sParams[group] == nil || sParams[group][label] == nil {
		return "", ErrParameterNotFound
	}

	return sParams[group][label].Get()
}

// Get all string values based on group
func GetStrings(group string) ([]string, error) {
	if sParams[group] == nil {
		return nil, ErrParameterNotFound
	}

	values := make([]string, len(sParams[group]))
	i := 0
	for _, param := range sParams[group] {
		if value, err := param.Get(); err == nil {
			values[i] = value
			i++
		} else {
			return nil, err
		}
	}

	return values, nil
}

// Get single float value based on group and label
func GetFloat(group, label string) (float64, error) {
	if fParams[group] == nil || fParams[group][label] == nil {
		return 0, ErrParameterNotFound
	}

	return fParams[group][label].Get()
}

// Get all float values based on group
func GetFloats(group string) ([]float64, error) {
	if fParams[group] == nil {
		return nil, ErrParameterNotFound
	}

	values := make([]float64, len(fParams[group]))
	i := 0
	for _, param := range fParams[group] {
		if value, err := param.Get(); err == nil {
			values[i] = value
			i++
		} else {
			return nil, err
		}
	}

	return values, nil
}

// Get single int value based on group and label
func GetInt(group, label string) (int, error) {
	if iParams[group] == nil || iParams[group][label] == nil {
		return 0, ErrParameterNotFound
	}

	return iParams[group][label].Get()
}

// Get all int values based on group
func GetInts(group string) ([]int, error) {
	if iParams[group] == nil {
		return nil, ErrParameterNotFound
	}

	values := make([]int, len(iParams[group]))
	i := 0
	for _, param := range iParams[group] {
		if value, err := param.Get(); err == nil {
			values[i] = value
			i++
		} else {
			return nil, err
		}
	}

	return values, nil
}
