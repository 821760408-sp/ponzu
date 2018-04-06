package admin

import (
	"net/http"
	"strings"
	"github.com/821760408-sp/ponzu/system/item"
	"github.com/821760408-sp/ponzu/management/format"
	"encoding/csv"
	"github.com/821760408-sp/ponzu/system/db"
)

func importHandler(res http.ResponseWriter, req *http.Request) {
	// /admin/contents/import?type=Faculty&format=csv
	if req.Method == "POST" {
		q := req.URL.Query()
		t := q.Get("type")
		f := strings.ToLower(q.Get("format"))

		if t == "" || f == "" {
			v, err := Error400()
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			res.WriteHeader(http.StatusBadRequest)
			_, err = res.Write(v)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

		}

		pt, ok := item.Types[t]
		if !ok {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		switch f {
		case "csv":
			_, ok := pt().(format.CSVFormattable)
			if !ok {
				res.WriteHeader(http.StatusBadRequest)
				return
			}

			//err := req.ParseMultipartForm(8 * 1024 * 1024)
			//if err != nil {
			//	res.WriteHeader(http.StatusInternalServerError)
			//	return
			//}

			//myCSV := req.Form.Get("myCSV")

			myCSV, _, err := req.FormFile("myCSV")

			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			defer myCSV.Close()
			r := csv.NewReader(myCSV)
			rec, err := r.ReadAll()

			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				return
			}

			// get content data and loop through creating a csv row per result
			bb := db.ContentAll(t)

			for _, line := range rec {}

			//importCSV(res, req, pt, fields)
		}
	}
}
