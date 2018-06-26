package stmt

// Option is a statement buffer option.
type Option func(*Stmt)

// AllowDollar is a statement buffer option to set allowing dollar strings (ie,
// $$text$$ or $tag$text$tag$).
func AllowDollar(enable bool) func(*Stmt) {
	return func(b *Stmt) {
		b.allowDollar = enable
	}
}

// AllowMultilineComments is a statement buffer option to set allowing multiline comments
// (ie, /* ... */).
func AllowMultilineComments(enable bool) func(*Stmt) {
	return func(b *Stmt) {
		b.allowMultilineComments = enable
	}
}

// AllowCComments is a statement buffer option to set allowing C-style comments
// (ie, // ...).
func AllowCComments(enable bool) func(*Stmt) {
	return func(b *Stmt) {
		b.allowCComments = enable
	}
}

// AllowHashComments is a statement buffer option to set allowing hash comments
// (ie, # ...).
func AllowHashComments(enable bool) func(*Stmt) {
	return func(b *Stmt) {
		b.allowHashComments = enable
	}
}
