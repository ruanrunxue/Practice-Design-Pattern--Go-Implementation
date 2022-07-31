package db

import (
	"math/rand"
	"time"
)

// 基于函数闭包的实现

type Next func(interface{}) error
type HasNext func() bool

func (t *Table) ClosureIterator() (HasNext, Next) {
	var records []record
	for _, r := range t.records {
		records = append(records, r)
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(records), func(i, j int) {
		records[i], records[j] = records[j], records[i]
	})
	size := len(records)
	cursor := 0
	hasNext := func() bool {
		return cursor < size
	}
	next := func(next interface{}) error {
		record := records[cursor]
		cursor++
		if err := record.convertByValue(next); err != nil {
			return err
		}
		return nil
	}
	return hasNext, next
}
