package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

var (
	deviceToken = "eec85466-0c0c-4e8c-acd3-09764ad9368f"
	pointValues = "FAI16-1-B:20.470;FAI16-2-B:20.473;FAI16-3-B:20.471;FAI17-1-B:9.866;FAI17-2-B:9.864;FAI17-3-B:9.864;FAI18-1-B:26.531;FAI18-2-B:26.534;FAI18-3-B:26.531;FAI23-1-B:15.273;FAI23-2-B:15.275;FAI23-3-B:15.272;FAI24-1-B:14.815;FAI24-2-B:14.817;FAI24-3-B:14.817;FAI35-1-B:0.465;FAI35-2-B:0.463;FAI35-3-B:0.465;FAI38-1-B:31.065;FAI38-2-B:31.068;FAI38-3-B:31.063;FAI100:9.804;FAI100-Y:7.581;FAI100-X:7.670;FAI99:0.023;FAI95:9.673;FAI95-Y:23.002;FAI95-X:7.670;FAI96:0.024;FAI98:8.622;FAI98-Y:15.287;FAI98-X:21.030;FAI97:0.010;FAI81-1:0.216;FAI81-2:0.237;FAI81-3:0.226;FAI81-4:0.222;FAI16-1-F:20.482;FAI16-2-F:20.491;FAI16-3-F:20.479;FAI17-1-F:9.867;FAI17-2-F:9.858;FAI17-3-F:9.865;FAI18-1-F:26.525;FAI18-2-F:26.529;FAI18-3-F:26.521;FAI23-1-F:15.260;FAI23-2-F:15.256;FAI23-3-F:15.254;FAI24-1-F:14.825;FAI24-2-F:14.824;FAI24-3-F:14.822;FAI35-1-F:0.482;FAI35-2-F:0.470;FAI35-3-F:0.468;FAI38-1-F:31.068;FAI38-2-F:31.070;FAI38-3-F:31.068;"
	qualified   = 1
	//barCode     = "FTTA31703CX2867HE101B17K17MT"
	barCode = ""

	url = "http://localhost/produce"
)

func main() {
	client := &http.Client{}

	requestBody := make(map[string]interface{})
	requestBody["device_token"] = deviceToken
	requestBody["point_value"] = pointValues
	requestBody["bar_code"] = barCode
	requestBody["qualified"] = qualified

	content, err := json.Marshal(requestBody)
	if err != nil {
		panic(err)
	}
	body := bytes.NewBuffer(content)

	req, err := http.NewRequest("POST", url, body)
	//req.Header.Add("If-None-Match", `W/"wyzzy"`)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)

}
