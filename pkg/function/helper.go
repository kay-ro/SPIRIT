package function

import "math"

// returns the maximum scope of all of the function scopes
func GetMaximumScope(functions ...*Function) *Scope {
	if len(functions) == 0 {
		return nil
	}

	maxS := &Scope{
		MinX: math.MaxFloat64,
		MinY: math.MaxFloat64,
		MaxX: -math.MaxFloat64,
		MaxY: -math.MaxFloat64,
	}

	for _, f := range functions {
		if f.Scope == nil {
			continue
		}

		maxS.MinX = min(maxS.MinX, f.Scope.MinX)
		maxS.MinY = min(maxS.MinY, f.Scope.MinY)
		maxS.MaxX = max(maxS.MaxX, f.Scope.MaxX)
		maxS.MaxY = max(maxS.MaxY, f.Scope.MaxY)
	}

	return maxS
}

func (s *Scope) CombineScope(s2 *Scope) {
	s.MinX = min(s.MinX, s2.MinX)
	s.MinY = min(s.MinY, s2.MinY)
	s.MaxX = max(s.MaxX, s2.MaxX)
	s.MaxY = max(s.MaxY, s2.MaxY)
}
