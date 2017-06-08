package model

// Permission is a bitwise access mask
type Permission int32

const (
	// CanRead translations on binded and back languages
	CanRead Permission = 1 << (iota + 1)
	// CanReadAll all translations
	CanReadAll
	// CanEdit translations on binded languages
	CanEdit
	// CanEditAll translations
	CanEditAll
	// CanAppend new messages
	CanAppend
	// CanDelete messages
	CanDelete

	// CanEverything is a total access
	CanEverything = CanRead | CanReadAll | CanEdit | CanEditAll | CanAppend | CanDelete
)
