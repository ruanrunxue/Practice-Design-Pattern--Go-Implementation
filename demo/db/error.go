package db

import "errors"

var (
	ErrPrimaryKeyConflict  = errors.New("primary key conflict")
	ErrRecordNotFound      = errors.New("record not found")
	ErrRecordTypeInvalid   = errors.New("record type invalid")
	ErrTableNotExist       = errors.New("table not exist")
	ErrTableAlreadyExist   = errors.New("table already exist")
	ErrTransactionNotBegin = errors.New("transaction not begin")
)
