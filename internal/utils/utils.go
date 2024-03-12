package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type ErrorStatus struct {
	Reason string `json:"reason"`
	Err    int    `json:"err"`
}

type OhlcvResponse struct {
	TokenName     string          `json:"tokenName"`
	Interval      string          `json:"interval"`
	VolumeArray   []interface{}   `json:"volumeArray"`
	FormattedOhlc [][]interface{} `json:"formattedOhlc"`
}

func formatDate(target map[string]interface{}) []int64 {
	var formattedDate []int64
	for date := range target {
		parsedTime, err := time.Parse(time.RFC3339, date)
		if err == nil {
			formattedDate = append(formattedDate, parsedTime.Unix())
		}
	}
	return formattedDate
}

func reshapeObject(target map[string]interface{}) [][]interface{} {
	var objValues [][]interface{}
	for _, values := range target {
		var innerSlice []interface{}
		for _, val := range values.(map[string]interface{}) {
			innerSlice = append(innerSlice, val)
		}
		objValues = append(objValues, innerSlice)
	}
	return objValues
}

func getVolumeArrayFromOhlcv(objectVals [][]interface{}) []interface{} {
	var volumeArray []interface{}
	for _, thing := range objectVals {
		volume := thing[len(thing)-1] // Assuming the volume is the last element
		volumeArray = append(volumeArray, volume)
	}
	return volumeArray
}

func zip(a []int64, b [][]interface{}) [][]interface{} {
	var zipped [][]interface{}
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	for i := 0; i < minLen; i++ {
		pair := []interface{}{a[i], b[i]}
		zipped = append(zipped, pair)
	}

	return zipped
}

func processAPIResponse(resp *http.Response) (interface{}, error) {
	if resp.StatusCode != http.StatusOK {
		errorRes := ErrorStatus{
			Reason: resp.Status,
			Err:    resp.StatusCode,
		}
		return nil, fmt.Errorf("API request failed: %v", errorRes)
	}

	var responseData map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		return nil, err
	}

	metaData, ok := responseData["Meta Data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("Meta Data missing in response")
	}

	tokenName := metaData["3. Digital Currency Name"].(string)
	interval := metaData["7. Interval"].(string)

	target := responseData[fmt.Sprintf("Time Series Crypto (%s)", interval)].(map[string]interface{})

	formattedDate := formatDate(target)
	objectValues := reshapeObject(target)
	volumeArray := getVolumeArrayFromOhlcv(objectValues)

	formattedData := OhlcvResponse{
		TokenName:     tokenName,
		Interval:      interval,
		FormattedOhlc: zip(formattedDate, objectValues),
		VolumeArray:   volumeArray,
	}
	fmt.Printf("Formatted data: %v\n", formattedData)
	return formattedData, nil
}
