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
		b.allowMc = enable
	}
}
