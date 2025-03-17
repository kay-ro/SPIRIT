package function

import "math"

// returns the maximum scope of all of the function scopes (needed for plotting)
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
		maxS.CombineScope(f.Scope)
	}

	return maxS
}

func (s *Scope) CombineScope(s2 *Scope) {
	s.MinX = min(s.MinX, s2.MinX)
	s.MinY = min(s.MinY, s2.MinY)
	s.MaxX = max(s.MaxX, s2.MaxX)
	s.MaxY = max(s.MaxY, s2.MaxY)
}
