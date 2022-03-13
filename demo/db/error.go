package db

import "errors"

var (
	ErrPrimaryKeyConflict  = errors.New("primary key conflict")
	ErrRecordNotFound      = errors.New("model not found")
	ErrRecordTypeInvalid   = errors.New("model type invalid")
	ErrTableNotExist       = errors.New("table not exist")
	ErrTableAlreadyExist   = errors.New("table already exist")
	ErrTransactionNotBegin = errors.New("transaction not begin")
	ErrDslInvalidGrammar   = errors.New("dsl expression invalid grammar")
)
