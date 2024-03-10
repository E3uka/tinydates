package tinydates

type User struct {
	Id       int    `json:id`
	Email    string `json:email`
	Password string `json:password`
	Name     string `json:name`
	Gender   string `json:gender`
	Age      int    `json:age`
}

type LoginRequest struct {
	Email    string `json:email`
	Password string `json:password`
}

type LoginResponse struct {
	Token string `json:token`
}

// GenericErrResponse is a generic error result return to the caller after an
// error is raised from an endpoint. The appropriate error reason should be
// returned to the caller.
type GenericErrResponse struct {
	Err string `json:omitempty`
}
