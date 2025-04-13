package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
)

// RequestXML represents the structure of incoming XML messages
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

// RequestJSON represents the structure of outgoing JSON API requests
type RequestJSON struct {
	Info           RequestInfo `json:"RequestInfo"`
	ParameterName  string      `json:"searchparametername"`
	ParameterValue string      `json:"searchparametervalue"`
}

// RequestInfo contains common request metadata
type RequestInfo struct {
	Stan        string `json:"requestId"`
	UserID      string `json:"userId,omitempty"`
	BaseNumber  string `json:"basenumber"`
	ChanelID    string `json:"chanelId"`
	RequestTime string `json:"requestTime"`
}

// RequestX2J transforms an XML message to a JSON API request
func RequestX2J(req []byte) ([]byte, error) {
	// Parse XML request
	xmlReq := &RequestXML{}
	if err := xml.Unmarshal(req, xmlReq); err != nil {
		return nil, fmt.Errorf("failed to parse XML request: %w", err)
	}

	// Transform to JSON format
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

	// Serialize to JSON
	result, err := json.Marshal(jsonReq)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize JSON request: %w", err)
	}

	return result, nil
}
