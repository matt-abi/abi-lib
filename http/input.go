package http

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ability-sh/abi-lib/dynamic"
	"github.com/ability-sh/abi-lib/json"
)

func GetInputData(r *http.Request, httpBodySize int64) interface{} {
	var inputData interface{} = nil
	ctype := r.Header.Get("Content-Type")

	if ctype == "" {
		ctype = r.Header.Get("content-type")
	}

	if strings.Contains(ctype, "multipart/form-data") {
		inputData = map[string]interface{}{}
		r.ParseMultipartForm(httpBodySize)
		if r.MultipartForm != nil {
			for key, values := range r.MultipartForm.Value {
				dynamic.Set(inputData, key, values[0])
			}
			for key, values := range r.MultipartForm.File {
				dynamic.Set(inputData, key, values[0])
			}
		}
	} else if strings.Contains(ctype, "json") {

		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		if err == nil {
			json.Unmarshal(b, &inputData)
		}

	} else {

		inputData = map[string]interface{}{}

		r.ParseForm()

		for key, values := range r.Form {
			dynamic.Set(inputData, key, values[0])
		}

	}
	return inputData
}
