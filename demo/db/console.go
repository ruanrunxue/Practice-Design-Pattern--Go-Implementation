package db

import (
	"bufio"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
	"strings"
)

/*
适配器模式
*/

// ConsoleRender 控制台db查询结果渲染接口
type ConsoleRender interface {
	Render() string
}

type Console struct {
	db Db
}

func NewConsole(db Db) *Console {
	return &Console{db: db}
}

func (c *Console) Start() {
	fmt.Println("welcome to Demo DB, enter exit to end!")
	fmt.Println("> please enter a sql expression:")
	fmt.Print("> ")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		sql := scanner.Text()
		if sql == "exit" {
			break
		}
		result, err := c.db.ExecSql(sql)
		if err == nil {
			c.Output(NewTableRender(result))
		} else {
			c.Output(NewErrorRender(err))
		}
		fmt.Println("> please enter a sql expression:")
		fmt.Print("> ")
	}
}

func (c *Console) Output(render ConsoleRender) {
	fmt.Println(render.Render())
}

type TableRender struct {
	result *SqlResult
}

func NewTableRender(result *SqlResult) *TableRender {
	return &TableRender{result: result}
}

func (t *TableRender) Render() string {
	vals := t.result.ToMap()
	var header []string
	var data []string
	for key, val := range vals {
		header = append(header, key)
		data = append(data, fmt.Sprintf("%v", val))
	}
	builder := &strings.Builder{}
	table := tablewriter.NewWriter(builder)
	table.SetHeader(header)
	table.Append(data)
	table.Render()
	return builder.String()
}

type ErrorRender struct {
	err error
}

func NewErrorRender(err error) *ErrorRender {
	return &ErrorRender{err: err}
}

func (e *ErrorRender) Render() string {
	return e.err.Error()
}
