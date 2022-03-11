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
	fmt.Println("enter exit to end.")
	fmt.Println("please enter a dsl expression:")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		dsl := scanner.Text()
		if dsl == "exit" {
			break
		}
		result, err := c.db.ExecDsl(dsl)
		if err == nil {
			c.Output(NewTableRender(result))
		} else {
			c.Output(NewErrorRender(err))
		}
		fmt.Println("please enter a dsl expression:")
	}
}

func (c *Console) Output(render ConsoleRender) {
	fmt.Println(render.Render())
}

type TableRender struct {
	result *DslResult
}

func NewTableRender(result *DslResult) *TableRender {
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
