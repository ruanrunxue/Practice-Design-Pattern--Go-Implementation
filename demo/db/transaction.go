package db

/*
命令模式
*/

// Command 执行数据库操作的命令接口
type Command interface {
	// Exec 执行insert、update、delete命令
	Exec() error
	// Undo 回滚命令
	Undo()
	// SetDb 设置关联的数据库
	setDb(db Db)
}

// Transaction Db事务实现，事务接口的调用顺序为begin -> exec -> exec > ... -> commit
type Transaction struct {
	name string
	db   Db
	cmds []Command
}

func NewTransaction(name string, db Db) *Transaction {
	return &Transaction{
		name: name,
		db:   db,
		cmds: nil,
	}
}

// Begin 开启一个事务
func (t *Transaction) Begin() {
	t.cmds = make([]Command, 0)
}

// Exec 在事务中执行命令，先缓存到cmds队列中，等commit时再执行
func (t *Transaction) Exec(cmd Command) error {
	if t.cmds == nil {
		return ErrTransactionNotBegin
	}
	cmd.setDb(t.db)
	t.cmds = append(t.cmds, cmd)
	return nil
}

// Commit 提交事务，执行队列中的命令，如果有命令失败，则回滚后返回错误
func (t *Transaction) Commit() error {
	history := &cmdHistory{history: make([]Command, 0, len(t.cmds))}
	for _, cmd := range t.cmds {
		if err := cmd.Exec(); err != nil {
			history.rollback()
			return err
		}
		history.add(cmd)
	}
	return nil
}

/*
备忘录模式
*/

// cmdHistory 命令执行历史
type cmdHistory struct {
	history []Command
}

func (c *cmdHistory) add(cmd Command) {
	c.history = append(c.history, cmd)
}

func (c *cmdHistory) rollback() {
	for i := len(c.history) - 1; i >= 0; i-- {
		c.history[i].Undo()
	}
}

// InsertCmd 插入命令
type InsertCmd struct {
	db         Db
	tableName  string
	primaryKey interface{}
	newRecord  interface{}
}

func NewInsertCmd(tableName string) *InsertCmd {
	return &InsertCmd{tableName: tableName}
}

func (i *InsertCmd) WithPrimaryKey(primaryKey interface{}) *InsertCmd {
	i.primaryKey = primaryKey
	return i
}

func (i *InsertCmd) WithRecord(record interface{}) *InsertCmd {
	i.newRecord = record
	return i
}

func (i *InsertCmd) Exec() error {
	return i.db.Insert(i.tableName, i.primaryKey, i.newRecord)
}

func (i *InsertCmd) Undo() {
	i.db.Delete(i.tableName, i.primaryKey)
}

func (i *InsertCmd) setDb(db Db) {
	i.db = db
}

// UpdateCmd 更新命令
type UpdateCmd struct {
	db         Db
	tableName  string
	primaryKey interface{}
	newRecord  interface{}
	oldRecord  interface{}
}

func NewUpdateCmd(tableName string) *UpdateCmd {
	return &UpdateCmd{tableName: tableName}
}

func (u *UpdateCmd) WithPrimaryKey(primaryKey interface{}) *UpdateCmd {
	u.primaryKey = primaryKey
	return u
}

func (u *UpdateCmd) WithRecord(record interface{}) *UpdateCmd {
	u.newRecord = record
	return u
}

func (u *UpdateCmd) Exec() error {
	if err := u.db.Query(u.tableName, u.primaryKey, u.oldRecord); err != nil {
		return err
	}
	return u.db.Update(u.tableName, u.primaryKey, u.newRecord)
}

func (u *UpdateCmd) Undo() {
	u.db.Update(u.tableName, u.primaryKey, u.oldRecord)
}

func (u *UpdateCmd) setDb(db Db) {
	u.db = db
}

// DeleteCmd 删除命令
type DeleteCmd struct {
	db         Db
	tableName  string
	primaryKey interface{}
	oldRecord  interface{}
}

func NewDeleteCmd(tableName string) *DeleteCmd {
	return &DeleteCmd{tableName: tableName}
}

func (d *DeleteCmd) WithPrimaryKey(primaryKey interface{}) *DeleteCmd {
	d.primaryKey = primaryKey
	return d
}

func (d *DeleteCmd) Exec() error {
	return d.db.Delete(d.tableName, d.primaryKey)
}

func (d *DeleteCmd) Undo() {
	d.db.Insert(d.tableName, d.primaryKey, d.oldRecord)
}

func (d *DeleteCmd) setDb(db Db) {
	d.db = db
}
