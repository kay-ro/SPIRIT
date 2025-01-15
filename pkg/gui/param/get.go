package param

// Get single string value based on group and label
func GetString(group, label string) (string, error) {
	if sParams[group] == nil {
		return "", ErrParameterNotFound
	}

	if p := sParams[group].GetParam(label); p != nil {
		return p.Get()
	}

	return "", ErrParameterNotFound
}

// Get all string values based on group
func GetStrings(group string) ([]string, error) {
	if sParams[group] == nil {
		return nil, ErrParameterNotFound
	}

	return sParams[group].GetValues()
}

// Get single float value based on group and label
func GetFloat(group, label string) (float64, error) {
	if fParams[group] == nil {
		return 0, ErrParameterNotFound
	}

	if p := fParams[group].GetParam(label); p != nil {
		return p.Get()
	}

	return 0, ErrParameterNotFound
}

// Get all float values based on group
func GetFloats(group string) ([]float64, error) {
	if fParams[group] == nil {
		return nil, ErrParameterNotFound
	}

	return fParams[group].GetValues()
}

// Get single int value based on group and label
func GetInt(group, label string) (int, error) {
	if iParams[group] == nil {
		return 0, ErrParameterNotFound
	}

	if p := iParams[group].GetParam(label); p != nil {
		return p.Get()
	}

	return 0, ErrParameterNotFound
}

// Get all int values based on group
func GetInts(group string) ([]int, error) {
	if iParams[group] == nil {
		return nil, ErrParameterNotFound
	}

	return iParams[group].GetValues()
}
