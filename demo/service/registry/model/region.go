package model

/*
享元模式
*/

// Region 值对象，每个服务都唯一属于一个Region
type Region struct {
	Id      string
	Name    string
	Country string
}

func NewRegion(id string) *Region {
	return &Region{Id: id}
}
