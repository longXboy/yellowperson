package schema

type MediaLog struct {
	Id         int64 `xorm:"pk autoincr"`
	Key        string
	MediaId    string
	Ip  	   string
	RefId 	int64
 	CreateTs   int
}
