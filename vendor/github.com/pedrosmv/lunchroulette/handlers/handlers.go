/*Package handlers consists on wrappers to the operation methods. This is done
due to the behaviour of HandleFunc method, it doesn't accept a function with
the w http.ResponseWriter, r *http.Request arguments, so they are encapsulated
by the wrapper that provides the mongo session */
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pedrosmv/lunchroulette/location"

	"goji.io/pat"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*CreateWrapper fetches the JSON with the new content and inserts it on the DB*/
func CreateWrapper(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		location := location.Location{}

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

		location := location.Location{}

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

		location := location.Location{}

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

		location := []location.Location{}

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
