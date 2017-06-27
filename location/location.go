package location

/*Location is the struct that will receive the values from a JSON to be stored
in the DB*/
type Location struct {
	ID      string `json:"id"`
	City    string `json:"city"`
	Country string `json:"country"`
	Street  string `json:"street"`
	Number  string `json:"number"`
}
