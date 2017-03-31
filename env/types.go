package env

// Vars is a variable handler interface.
type Vars interface {
	// Set sets a variable.
	Set(string, interface{}) error

	// Unset unsets a variable.
	Unset(string)
}

// Pvars is a pretty variable handler interface.
type Pvars interface {
}
