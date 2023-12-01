package requests

// Result is the type used for returning and propagating errors.
//
// It is an enum with the variants, Ok(T), representing success and containing a value,
// and Err(E), representing error and containing an error value.
type Result[T any] struct {
	t T
	e error
}

// Err equal to generate a Result[T]{E: err}
func Err[T any](err error) Result[T] {
	return Result[T]{e: err}
}

// Ok equal to generate a Result[T]{T: t}
func Ok[T any](t T) Result[T] {
	return Result[T]{t: t}
}

// Err Converts from Result<T, E> to Option<E>.
//
// Converts self into an Option<E>, consuming self, and discarding the success value, if any.
func (r Result[T]) Err() error {
	return r.e
}

func (r Result[T]) Ok() *T {
	if r.e != nil {
		return nil
	}
	return &r.t
}

// Unwrap Returns the contained Ok value, consuming the self value.
//
// Because this function may panic, its use is generally discouraged.
// Instead, prefer to use pattern matching and handle the Err case explicitly,
// or call unwrap_or, unwrap_or_else, or unwrap_or_default.
//
// Panics: Panics if the value is an Err, with a panic message provided by the Err’s value.
func (r Result[T]) Unwrap() T {
	if r.e != nil {
		panic(r.e)
	}
	return r.t
}

// UnwrapOr Returns the contained Ok value or a provided default.
//
// Arguments passed to unwrap_or are eagerly evaluated; if you are passing the result of a function call,
// it is recommended to use unwrap_or_else, which is lazily evaluated.
func (r Result[T]) UnwrapOr(defaultValue T) T {
	if r.e != nil {
		return defaultValue
	}
	return r.t
}

// AndThen Calls op if the result is Ok, otherwise returns the Err value of self.
//
// This function can be used for control flow based on Result values.
func andThen[T, U any](r Result[T], op func(data T) Result[U]) Result[U] {
	if r.e != nil {
		return Result[U]{e: r.e}
	}
	return op(r.t)
}
