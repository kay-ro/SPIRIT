package option

type Option[T any] struct {
	isSome bool
	value  *T
}

func Some[T any](value *T) Option[T] {
	return Option[T]{
		isSome: true,
		value:  value,
	}
}

func None[T any]() Option[T] {
	return Option[T]{
		isSome: false,
		value:  nil,
	}
}

func (o Option[T]) IsSome() bool {
	return o.isSome
}

func (o Option[T]) Unwrap() T {
	if o.isSome {
		return *o.value
	} else {
		panic("Can not get value from None Option")
	}
}

func (o Option[T]) UnwrapOr(altValue T) T {
	if o.isSome {
		return *o.value
	} else {
		return altValue
	}
}
