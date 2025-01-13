package function

// returns the maximum scope of all of the function scopes
func GetMaximumScope(functions ...*Function) *Scope {
	if len(functions) == 0 {
		return nil
	}

	c := *functions[0].Scope
	maxS := &c

	for _, f := range functions[1:] {
		maxS.MinX = min(maxS.MinX, f.Scope.MinX)
		maxS.MinY = min(maxS.MinY, f.Scope.MinY)
		maxS.MaxX = max(maxS.MaxX, f.Scope.MaxX)
		maxS.MaxY = max(maxS.MaxY, f.Scope.MaxY)
	}

	return maxS
}
