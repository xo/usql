package handler

import "github.com/xo/usql/text"

// branch is the state of one block of \if, as psql's ifState.
type branch int

// The states of a block.
const (
	// branchTrue is a branch that runs, before the \else of its block.
	branchTrue branch = iota
	// branchFalse is a branch that does not run, in a block where no
	// branch ran yet, so that an \elif or the \else can still run.
	branchFalse
	// branchIgnored is a branch that does not run, because a branch of its
	// block before it ran, or because the block is in a branch that does
	// not run. No later branch of the block runs.
	branchIgnored
	// branchElseTrue is an \else that runs.
	branchElseTrue
	// branchElseFalse is an \else that does not run.
	branchElseFalse
)

// conds is psql's stack of conditional blocks, with the innermost block last.
// Each input has a stack of its own: the terminal, the text of -c, and each
// file of -f and \i.
type conds []branch

// active reports whether the current branch runs. Outside a block,
// everything runs.
func (c conds) active() bool {
	if len(c) == 0 {
		return true
	}
	b := c[len(c)-1]
	return b == branchTrue || b == branchElseTrue
}

// begin starts a block, as \if does. It calls eval for the expression only
// when the block starts in a branch that runs, so that a branch that does
// not run expands no variable and runs no command between backticks. An
// expression that is not a boolean is an error, and the branch does not
// run.
func (c *conds) begin(eval func() (bool, error)) error {
	if !c.active() {
		*c = append(*c, branchIgnored)
		return nil
	}
	ok, err := eval()
	b := branchTrue
	if err != nil || !ok {
		b = branchFalse
	}
	*c = append(*c, b)
	return err
}

// elif starts the next branch of the block, as \elif does. It calls eval only
// when no branch of the block ran yet.
func (c *conds) elif(eval func() (bool, error)) error {
	if len(*c) == 0 {
		return text.ErrElifNoMatchingIf
	}
	top := &(*c)[len(*c)-1]
	switch *top {
	case branchTrue:
		*top = branchIgnored
	case branchFalse:
		ok, err := eval()
		if err != nil || !ok {
			return err
		}
		*top = branchTrue
	case branchElseTrue, branchElseFalse:
		return text.ErrElifAfterElse
	}
	return nil
}

// els starts the last branch of the block, as \else does. It runs when no
// branch of the block ran.
func (c *conds) els() error {
	if len(*c) == 0 {
		return text.ErrElseNoMatchingIf
	}
	top := &(*c)[len(*c)-1]
	switch *top {
	case branchFalse:
		*top = branchElseTrue
	case branchTrue, branchIgnored:
		*top = branchElseFalse
	case branchElseTrue, branchElseFalse:
		return text.ErrElseAfterElse
	}
	return nil
}

// isConditional reports whether the meta command name is \if, \elif, \else or
// \endif, which run in every branch to keep track of the blocks.
func isConditional(name string) bool {
	switch name {
	case "if", "elif", "else", "endif":
		return true
	}
	return false
}

// end ends the block, as \endif does.
func (c *conds) end() error {
	if len(*c) == 0 {
		return text.ErrEndIfNoMatchingIf
	}
	*c = (*c)[:len(*c)-1]
	return nil
}
