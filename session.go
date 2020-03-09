package gin_session

//Session 接口
type Session interface{
	Get(string) (string, bool)
	Set(string, string)
  	GetStruct(string, interface{})error
	SetStruct(string, interface{})error
}

