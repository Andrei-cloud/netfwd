package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
)

// ResponseXML represents the structure of outgoing XML responses
type ResponseXML struct {
	XMLName        xml.Name      `xml:"XML"`
	MessageType    string        `xml:"MessageType"`
	ProcCode       string        `xml:"ProcCode"`
	Stan           string        `xml:"STAN"`
	RequestTime    string        `xml:"LocalTxnDtTime"`
	ChanelID       string        `xml:"DeliveryChannelCtrlID"`
	ParameterName  string        `xml:"PName"`
	ParameterValue string        `xml:"PValue"`
	ActCode        string        `xml:"ActCode"`
	ActDescription string        `xml:"ActDescription"`
	TotalnoofTrans int           `xml:"TotalnoofTrans"`
	Customers      CustomersList `xml:"Customers"`
	RefNum         string        `xml:"REFNUM"`
}

// CustomersList represents a collection of customer records
type CustomersList struct {
	Records []Record `xml:"Record"`
}

// Record represents a customer record in the XML response
type Record struct {
	Name                   string `xml:"Name"`
	FirstName              string `xml:"FirstName"`
	MiddleName             string `xml:"MiddleName"`
	LastName               string `xml:"LastName"`
	BaseNumber             string `xml:"BaseNumber"`
	Nationality            string `xml:"Nationality"`
	PoBox                  string `xml:"PoBox"`
	Address                string `xml:"Address"`
	City                   string `xml:"City"`
	Country                string `xml:"Country"`
	Email                  string `xml:"Email"`
	CardOnlyCustomer       string `xml:"CardOnlyCustomer"`
	SMSMobile              string `xml:"SMSMobile"`
	SMSLang                string `xml:"SMSLang"`
	SMSNationalID          string `xml:"SMSNationalID"`
	SMSPassportNo          string `xml:"SMSPassportNo"`
	SegmentCode            string `xml:"SegmentCode"`
	SegmentDesc            string `xml:"SegmentDesc"`
	QID                    string `xml:"QID"`
	QIDExpiryDate          string `xml:"QIDExpiryDate"`
	PassportNo             string `xml:"PassportNo"`
	PassportExpiryDate     string `xml:"PassportExpiryDate"`
	CompanyRegNo           string `xml:"CompanyRegNo"`
	CompanyRegNoExpiryDate string `xml:"CompanyRegNoExpiryDate"`
	LOB                    string `xml:"LOB"`
	DOB                    string `xml:"DOB"`
	CustTypeFlag           string `xml:"CustTypeFlag"`
}

// ResponseJSON represents the structure of incoming JSON API responses
type ResponseJSON struct {
	Info      RequestInfo `json:"RequestInfo"`
	Customers []Details   `json:"CustomerDetails"`
}

// Details represents a customer record in the JSON response
type Details struct {
	Qid                      string      `json:"QID"`
	Baseno                   string      `json:"BASENO"`
	Crno                     interface{} `json:"CRNO"`
	Passportno               string      `json:"PASSPORTNO"`
	Mobileno                 string      `json:"MOBILENO"`
	Emailid                  string      `json:"EMAILID"`
	GUID                     string      `json:"GUID"`
	FirstName                string      `json:"FirstName"`
	LastName                 string      `json:"LastName"`
	IsBlacklisted            bool        `json:"IsBlacklisted"`
	IsQANationalityWithdrawn bool        `json:"IsQANationalityWithdrawn"`
}

// ResponseJ2X transforms a JSON API response to an XML message
func ResponseJ2X(res []byte) ([]byte, error) {
	// Parse JSON response
	jsonRes := &ResponseJSON{}
	if err := json.Unmarshal(res, jsonRes); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Transform to XML format
	xmlRes := &ResponseXML{
		MessageType:    "1",
		ProcCode:       "CSNQ",
		Stan:           jsonRes.Info.Stan,
		RequestTime:    jsonRes.Info.RequestTime,
		ChanelID:       "ATM",
		ParameterName:  "ACCOUNTNUMBER",
		ParameterValue: jsonRes.Info.BaseNumber,
		ActCode:        "0",
		ActDescription: "Success",
		TotalnoofTrans: len(jsonRes.Customers),
		RefNum:         jsonRes.Info.UserID,
	}

	// Transform customer records
	records := make([]Record, 0, len(jsonRes.Customers))
	for _, c := range jsonRes.Customers {
		record := Record{
			Name:       c.FirstName + " " + c.LastName,
			FirstName:  c.FirstName,
			LastName:   c.LastName,
			BaseNumber: c.Baseno,
			QID:        c.Qid,
			PassportNo: c.Passportno,
			SMSMobile:  c.Mobileno,
			Email:      c.Emailid,
		}
		records = append(records, record)
	}
	xmlRes.Customers.Records = records

	// Serialize to XML
	result, err := xml.Marshal(xmlRes)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize XML response: %w", err)
	}

	return result, nil
}
