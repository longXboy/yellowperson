package schema

type AccLog struct {
	Id         int64 `xorm:"pk autoincr"`
	IP        string
	URI    string
	Method		string
	UA  	   string
	StatusCode  int  
	ContentLength  int
	ResponseTs int
	CreateTs 	int
}
