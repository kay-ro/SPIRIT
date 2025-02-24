package param

// SetString sets a string value for a specific group and label
func SetString(group, label, value string) error {
	if sParams[group] == nil || !sParams[group].Check(label) {
		return ErrParameterNotFound
	}

	return sParams[group].Set(label, value)
}

// SetStrings sets string values for a specific group
func SetStrings(group string, values []string) error {
	if sParams[group] == nil {
		return ErrParameterNotFound
	}

	return sParams[group].SetAll(values)
}

// SetFloat sets a float value for a specific group and label
func SetFloat(group, label string, value float64) error {
	if fParams[group] == nil || !fParams[group].Check(label) {
		return ErrParameterNotFound
	}

	return fParams[group].Set(label, value)
}

// SetFloats sets float values for a specific group
func SetFloats(group string, values []float64) error {
	if fParams[group] == nil {
		return ErrParameterNotFound
	}

	return fParams[group].SetAll(values)
}

// SetInt sets an int value for a specific group and label
func SetInt(group, label string, value int) error {
	if iParams[group] == nil || !iParams[group].Check(label) {
		return ErrParameterNotFound
	}

	return iParams[group].Set(label, value)
}

// SetInts sets int values for a specific group
func SetInts(group string, values []int) error {
	if iParams[group] == nil {
		return ErrParameterNotFound
	}

	return iParams[group].SetAll(values)
}
