package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
)

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

type CustomersList struct {
	Records []Record `xml:"Record"`
}
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

type ResponseJSON struct {
	Info      RequestInfo `json:"RequestInfo"`
	Customers []Details   `json:"CustomerDetails"`
}

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

func ResponseJ2X(res []byte) ([]byte, error) {
	jsonRes := &ResponseJSON{}
	err := json.Unmarshal(res, jsonRes)
	if nil != err {
		return []byte{}, fmt.Errorf("requst xml unmarshal: %w", err)
	}

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

	records := make([]Record, 0)

	for _, c := range jsonRes.Customers {
		cust := Record{}
		cust.Name = c.FirstName + " " + c.LastName
		cust.FirstName = c.FirstName
		cust.LastName = c.LastName
		cust.BaseNumber = c.Baseno
		cust.QID = c.Qid
		cust.PassportNo = c.Passportno
		cust.SMSMobile = c.Mobileno
		cust.Email = c.Emailid

		records = append(records, cust)
	}

	xmlRes.Customers.Records = records

	result, err := xml.Marshal(xmlRes)
	if err != nil {
		return []byte{}, fmt.Errorf("requst xml unmarshal: %w", err)
	}

	return result, nil
}
