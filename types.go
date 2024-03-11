package tinydates

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Gender   string `json:"gender"`
	Age      int    `json:"age"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type DiscoveredUser struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Gender string `json:"gender"`
	Age    int    `json:"age"`
}

type DiscoverResponse struct {
	Results []DiscoveredUser `json:"results"`
}

type SwipeRequest struct {
	SwiperId int  `json:"swiperId"`
	SwipeeId int  `json:"swipeeId"`
	Decision bool `json:"decision"`
}

type SwipeResponse struct {
	Matched bool `json:"matched"`
	MatchId int  `json:"matchID,omitempty"`
}

// GenericErrResponse is a generic error result return to the caller after an
// error is raised from an endpoint. The appropriate error reason should be
// returned to the caller.
type GenericErrResponse struct {
	Err string `json:"error,omitempty"`
}
