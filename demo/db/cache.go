package db

import "sync"

/*
代理模式
*/

// CacheProxy Db缓存代理
type CacheProxy struct {
	db    Db
	cache sync.Map // key为tableName，value为sync.Map[key: primaryId, value: interface{}]
	hit   int
	miss  int
}

func NewCacheProxy(db Db) *CacheProxy {
	return &CacheProxy{
		db:    db,
		cache: sync.Map{},
		hit:   0,
		miss:  0,
	}
}

func (c *CacheProxy) Hit() int {
	return c.hit
}

func (c *CacheProxy) Miss() int {
	return c.miss
}

func (c *CacheProxy) CreateTable(t *Table) error {
	if err := c.db.CreateTable(t); err != nil {
		return err
	}
	c.cache.Store(t.Name(), &sync.Map{})
	return nil
}

func (c *CacheProxy) CreateTableIfNotExist(t *Table) error {
	if _, ok := c.cache.Load(t.Name()); !ok {
		c.cache.Store(t.Name(), t)
	}
	return c.db.CreateTableIfNotExist(t)
}

func (c *CacheProxy) DeleteTable(tableName string) error {
	c.cache.Delete(tableName)
	return c.db.DeleteTable(tableName)
}

func (c *CacheProxy) Query(tableName string, primaryKey interface{}, result interface{}) error {
	cache, ok := c.cache.Load(tableName)
	if ok {
		if record, ok := cache.(*sync.Map).Load(primaryKey); ok {
			c.hit++
			result = record
			return nil
		}
	}
	c.miss++
	if err := c.db.Query(tableName, primaryKey, result); err != nil {
		return err
	}
	cache.(*sync.Map).Store(primaryKey, result)
	return nil
}

func (c *CacheProxy) QueryByVisitor(tableName string, visitor TableVisitor) ([]interface{}, error) {
	return c.db.QueryByVisitor(tableName, visitor)
}

func (c *CacheProxy) Insert(tableName string, primaryKey interface{}, record interface{}) error {
	if err := c.db.Insert(tableName, primaryKey, record); err != nil {
		return err
	}
	cache, ok := c.cache.Load(tableName)
	if !ok {
		return nil
	}
	cache.(*sync.Map).Store(primaryKey, record)
	return nil
}

func (c *CacheProxy) Update(tableName string, primaryKey interface{}, record interface{}) error {
	if err := c.db.Update(tableName, primaryKey, record); err != nil {
		return err
	}
	cache, ok := c.cache.Load(tableName)
	if !ok {
		return nil
	}
	cache.(*sync.Map).Store(primaryKey, record)
	return nil
}

func (c *CacheProxy) Delete(tableName string, primaryKey interface{}) error {
	if err := c.db.Delete(tableName, primaryKey); err != nil {
		return err
	}
	cache, ok := c.cache.Load(tableName)
	if !ok {
		return nil
	}
	cache.(*sync.Map).Delete(primaryKey)
	return nil
}
