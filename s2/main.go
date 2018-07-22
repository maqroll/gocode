package main

import "github.com/golang/geo/s2"
import "net/http"
import "log"
import "fmt"
import "strconv"
import "encoding/json"

type result struct {
    Lat float64 `json:"Lat"`
    Lon float64 `json:"Lon"`
}

func s2toLonLat(w http.ResponseWriter, r *http.Request){
	strCell := r.FormValue("cell")
	fmt.Println(strCell)
	cell,_ := strconv.ParseUint(strCell,0,64)
	ll := s2.CellID(cell).LatLng()

	res := result{
		Lat: ll.Lat.Degrees(),
		Lon: ll.Lng.Degrees() }
	
	json.NewEncoder(w).Encode(res)
}

func handleRequests() {
    http.HandleFunc("/", s2toLonLat)
    log.Fatal(http.ListenAndServe(":8081", nil))
}

func main() {
    handleRequests()
}