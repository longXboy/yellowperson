package schema

type ErrLog struct {
	Id         int64 `xorm:"pk autoincr"`
	ErrType        string
	ErrContent    string
	MediaId		string
	Ip  	   string
	CreateTs 	int
}
