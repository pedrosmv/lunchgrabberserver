package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"goji.io/pat"

	goji "goji.io"

	"github.com/rs/cors"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Location struct {
	ID      string `json:"id"`
	City    string `json:"city"`
	Country string `json:"country"`
	Street  string `json:"street"`
	Number  string `json:"number"`
}

/*Index is a struct from the mongo package that groups settings for the DB. This
function makes sure that all of them are true */
func Index(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()

	context := session.DB("store").C("locations")

	index := mgo.Index{
		Key:        []string{"id"}, //Index key fields; prefix name with dash (-) for descending order
		Unique:     true,           //Prevent two documents from having the same index key
		DropDups:   true,           //Drop documents with the same index key as a previously indexed one
		Background: true,           //Build index in background and return immediately
		Sparse:     true,           //Only index documents containing the Key fields
	}

	err := context.EnsureIndex(index)
	if err != nil {
		fmt.Println(err)
	}
}

func determineListenAddress() (string, error) {
	port := os.Getenv("PORT")
	if port == "" {
		return "", fmt.Errorf("$PORT not set")
	}
	return ":" + port, nil
}

func main() {

	addr, err := determineListenAddress()

	session, err := mgo.Dial("localhost")
	if err != nil {
		fmt.Println(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	Index(session)

	multiplex := goji.NewMux()
	multiplex.HandleFunc(pat.Post("/locations"), CreateWrapper(session))
	multiplex.HandleFunc(pat.Get("/locations/:city"), FetchAll(session))
	multiplex.HandleFunc(pat.Get("/locations/:id"), ReadWrapper(session))
	multiplex.HandleFunc(pat.Put("/locations/:id"), UpdateWrapper(session))
	multiplex.HandleFunc(pat.Delete("/locations/:id"), DeleteWrapper(session))
	handler := cors.Default().Handler(multiplex)
	http.ListenAndServe(addr, handler)
}

/*CreateWrapper fetches the JSON with the new content and inserts it on the DB*/
func CreateWrapper(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		location := Location{}

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&location)
		if err != nil {
			fmt.Println(err)
		}

		context := session.DB("store").C("locations")

		err = context.Insert(location)
		fmt.Println(location)
		if err != nil {
			fmt.Println(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", r.URL.Path+"/"+location.ID)
		w.WriteHeader(http.StatusCreated)
	}
}

/*ReadWrapper will look for the object with the ID received from the request and
return it*/
func ReadWrapper(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		id := pat.Param(r, "id")

		context := session.DB("store").C("locations")

		location := Location{}

		// err := context.Find(bson.M{"id": id}).One(&location)
		err := context.Find(bson.M{"id": id}).One(&location)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(location)
		response, err := json.MarshalIndent(location, "", "  ")
		if err != nil {
			fmt.Println(err)
		}
		w.Write(response)
	}
}

/*UpdateWrapper works similarly to ReadWrapper, it will look for the entry with the
given ID and update it with the JSON received*/
func UpdateWrapper(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		id := pat.Param(r, "id")

		location := Location{}

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&location)
		if err != nil {
			fmt.Println(err)
		}

		context := session.DB("store").C("locations")
		err = context.Update(bson.M{"id": id}, &location)
		if err != nil {
			fmt.Println(err)
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

/*DeleteWrapper will delete the entry that matches with the ID given */
func DeleteWrapper(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		id := pat.Param(r, "id")
		context := session.DB("store").C("locations")
		err := context.Remove(bson.M{"id": id})
		if err != nil {
			fmt.Println(err)
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

/*FetchAll is responsible for searching the database for locations of a given city */
func FetchAll(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		city := pat.Param(r, "city")

		context := session.DB("store").C("locations")

		location := []Location{}

		err := context.Find(bson.M{"city": city}).All(&location)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(location)
		response, err := json.MarshalIndent(location, "", "  ")
		if err != nil {
			fmt.Println(err)
		}
		w.Write(response)
	}
}
