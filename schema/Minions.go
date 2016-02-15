package schema

type Minions struct {
	Id         int64 `xorm:"pk autoincr"`
	CreateTs 	int
}
