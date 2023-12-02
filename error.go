package requests

// Result is the type used for returning and propagating errors.
//
// It is an enum with the variants, Ok(T), representing success and containing a value,
// and Err(E), representing error and containing an error value.
type Result[T any] struct {
	t T
	e error
}

// Err generate a Result[T]{E: err}
func Err[T any](err error) Result[T] {
	return Result[T]{e: err}
}

// Ok generate a Result[T]{T: t}
func Ok[T any](t T) Result[T] {
	return Result[T]{t: t}
}

// Err Converts from Result<T, E> to E
//
// Converts self into an E, consuming self, and discarding the success value, if any.
func (r Result[T]) Err() error {
	return r.e
}

func (r Result[T]) Value() T {
	return r.t
}

func (r Result[T]) IsErr() bool {
	return r.e != nil
}

func (r Result[T]) IsOk() bool {
	return r.e == nil
}

func (r Result[T]) Unpack() (T, error) {
	return r.t, r.e
}

// Or Returns the contained Ok value or a provided default.
//
// Arguments passed to unwrap_or are eagerly evaluated; if you are passing the result of a function call,
// it is recommended to use unwrap_or_else, which is lazily evaluated.
func (r Result[T]) Or(defaultValue T) T {
	if r.e != nil {
		return defaultValue
	}
	return r.t
}

func (r Result[T]) OrElse(fallback func(data T) Result[T]) Result[T] {
	if r.e != nil {
		return Err[T](r.e)
	}
	return fallback(r.t)
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

func Or[T, U any](r Result[T], defaultValue U, f func(T) U) U {
	if r.e != nil {
		return defaultValue
	}
	return f(r.t)
}

// Then Calls op if the result is Ok, otherwise returns the Err value of self.
//
// This function can be used for control flow based on Result values.
func Then[T, U any](r Result[T], op func(data T) Result[U]) Result[U] {
	if r.e != nil {
		return Err[U](r.e)
	}
	return op(r.t)
}
