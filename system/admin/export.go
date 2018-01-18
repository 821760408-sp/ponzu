package admin

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/821760408-sp/ponzu/management/format"
	"github.com/821760408-sp/ponzu/system/api"
	"github.com/821760408-sp/ponzu/system/db"
	"github.com/821760408-sp/ponzu/system/item"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func exportHandler(res http.ResponseWriter, req *http.Request) {
	// /admin/contents/export?type=Blogpost&format=csv
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
		csv, ok := pt().(format.CSVFormattable)
		if !ok {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		fields := csv.FormatCSV()
		exportCSV(res, req, pt, fields)

	case "json":
		json, ok := pt().(format.JSONFormattable)
		if !ok {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		exportJSON(res, req)

	default:
		res.WriteHeader(http.StatusBadRequest)
		return
	}
}

func exportCSV(res http.ResponseWriter, req *http.Request, pt func() interface{}, fields []string) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "exportcsv-")
	if err != nil {
		log.Println("Failed to create tmp file for CSV export:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = os.Chmod(tmpFile.Name(), 0666)
	if err != nil {
		log.Println("chmod err:", err)
	}

	csvBuf := csv.NewWriter(tmpFile)

	t := req.URL.Query().Get("type")

	// get content data and loop through creating a csv row per result
	bb := db.ContentAll(t)

	// add field names to first row
	err = csvBuf.Write(fields)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		log.Println("Failed to write column headers:", fields)
		return
	}

	for row := range bb {
		// unmarshal data and loop over fields
		rowBuf := []string{}

		for _, col := range fields {
			// pull out each field as the column value
			result := gjson.GetBytes(bb[row], col)

			// append it to the buffer
			rowBuf = append(rowBuf, result.String())
		}

		// write row to csv
		err := csvBuf.Write(rowBuf)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			log.Println("Failed to write column headers:", fields)
			return
		}
	}

	csvBuf.Flush()

	// write the buffer to a content-disposition response
	fi, err := tmpFile.Stat()
	if err != nil {
		log.Println("Failed to read tmp file info for CSV export:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = tmpFile.Close()
	if err != nil {
		log.Println("Failed to close tmp file for CSV export:", err)
	}

	ts := time.Now().Unix()
	disposition := `attachment; filename="export-%s-%d.csv"`

	res.Header().Set("Content-Type", "text/csv")
	res.Header().Set("Content-Disposition", fmt.Sprintf(disposition, t, ts))
	res.Header().Set("Content-Length", fmt.Sprintf("%d", int(fi.Size())))

	http.ServeFile(res, req, tmpFile.Name())

	err = os.Remove(tmpFile.Name())
	if err != nil {
		log.Println("Failed to remove tmp file for CSV export:", err)
	}
}

func exportJSON(res http.ResponseWriter, req *http.Request) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "exportjson-")
	if err != nil {
		log.Println("Failed to create tmp file for JSON export:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = os.Chmod(tmpFile.Name(), 0666)
	if err != nil {
		log.Println("chmod err:", err)
	}

	jsonBuf := bufio.NewWriter(tmpFile)

	t := req.URL.Query().Get("type")

	// get json data ("system/api/handlers")
	it, ok := item.Types[t]
	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	opts := db.QueryOptions{
		Count:  -1,     // 10 default, -1 is all
		Offset: 0,      // 0 default
		Order:  "desc", // DESC default
	}

	_, bb := db.Query(t+"__sorted", opts)
	var result = []json.RawMessage{}
	for i := range bb {
		result = append(result, bb[i])
	}

	j, err := api.FmtJSON(result...)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	j, err = api.Omit(res, req, it(), j)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = jsonBuf.Write(j)

	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		log.Println("Failed to write JSON file.")
		return
	}

	jsonBuf.Flush()

	// write the buffer to a content-disposition response
	fi, err := tmpFile.Stat()
	if err != nil {
		log.Println("Failed to read tmp file info for JSON export:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = tmpFile.Close()
	if err != nil {
		log.Println("Failed to close tmp file for CSV export:", err)
	}

	ts := time.Now().Unix()
	disposition := `attachment; filename="export-%s-%d.json"`

	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Content-Disposition", fmt.Sprintf(disposition, t, ts))
	res.Header().Set("Content-Length", fmt.Sprintf("%d", int(fi.Size())))

	http.ServeFile(res, req, tmpFile.Name())

	err = os.Remove(tmpFile.Name())
	if err != nil {
		log.Println("Failed to remove tmp file for JSON export:", err)
	}
}
