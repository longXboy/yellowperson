package schema

type Manage struct {
	UserId     string
	OpenId     string
	PassCode   string
	AccessCode string
	ExpiredTs  int
	Times      int
}
