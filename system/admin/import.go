package admin

import (
	"net/http"
	"strings"
	"github.com/821760408-sp/ponzu/system/item"
	"github.com/821760408-sp/ponzu/management/format"
	//"encoding/csv"
)

func importHandler(res http.ResponseWriter, req *http.Request) {
	// /admin/contents/import?type=Faculty&format=csv
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

		//importCSV(res, req, pt, fields)
	}
}
