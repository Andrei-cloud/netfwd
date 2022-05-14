package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
)

type RequestXML struct {
	XMLName        xml.Name `xml:"XML"`
	Text           string   `xml:",chardata"`
	MessageType    string   `xml:"MessageType"`
	ProcCode       string   `xml:"ProcCode"`
	RefNum         string   `xml:"REFNUM"`
	Stan           string   `xml:"STAN"`
	RequestTime    string   `xml:"LocalTxnDtTime"`
	ChanelID       string   `xml:"DeliveryChannelCtrlID"`
	ParameterName  string   `xml:"PName"`
	ParameterValue string   `xml:"PValue"`
}

type RequestJSON struct {
	Info           RequestInfo `json:"RequestInfo"`
	ParameterName  string      `json:"searchparametername"`
	ParameterValue string      `json:"searchparametervalue"`
}

type RequestInfo struct {
	Stan        string `json:"requestId"`
	UserID      string `json:"userId,omitempty"`
	BaseNumber  string `json:"basenumber"`
	ChanelID    string `json:"chanelId"`
	RequestTime string `json:"requestTime"`
}

func RequestX2J(req []byte) ([]byte, error) {
	xmlReq := &RequestXML{}

	err := xml.Unmarshal(req, xmlReq)
	if err != nil {
		return []byte{}, fmt.Errorf("request xml unmarshal: %w", err)
	}

	jsonReq := &RequestJSON{
		Info: RequestInfo{
			Stan:        xmlReq.Stan,
			UserID:      xmlReq.RefNum,
			BaseNumber:  xmlReq.ParameterValue,
			ChanelID:    xmlReq.ChanelID,
			RequestTime: xmlReq.RequestTime,
		},
		ParameterName:  "Baseno",
		ParameterValue: xmlReq.ParameterValue,
	}
	result, err := json.Marshal(jsonReq)
	if nil != err {
		return []byte{}, fmt.Errorf("request json marshal: %w", err)
	}
	return result, nil
}
