package env

// Vars is a map of variables.
type Vars map[string]string

// Set sets a variable name.
func (v Vars) Set(name, value string) {
	v[name] = value
}

// Unset unsets a variable name.
func (v Vars) Unset(name string) {
	delete(v, name)
}

// All returns all variables as a map.
func (v Vars) All() map[string]string {
	return map[string]string(v)
}

// Pvars is a map
type Pvars interface{}

var vars Vars

func init() {
	vars = make(Vars)
}

// Set sets a variable.
func Set(name, value string) {
	vars.Set(name, value)
}

// Unset unsets a variable.
func Unset(name string) {
	vars.Unset(name)
}

// All returns all variables.
func All() map[string]string {
	return vars
}
