package main

import (
	"reflect"
	"testing"
)

func TestResponseJ2X(t *testing.T) {
	tests := []struct {
		name    string
		req     []byte
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"Response",
			[]byte(
				`{"RequestInfo":{"requestId":"0220000245250","userId":"256557","basenumber":"157336","chanelId":"ATM","requestTime":"2203221157"},"CustomerDetails":[{"QID":"273XXXXXXXX","BASENO":"157336","CRNO":null,"PASSPORTNO":"XXXXXXXX","MOBILENO":"","EMAILID":"example@example.com","GUID":"7c7f7a47-f236-ea11-9132-00505685b1c3","FirstName":"IVAN","LastName":"IVANOV","IsBlacklisted":false,"IsQANationalityWithdrawn":false}]}`,
			),
			[]byte(
				`<XML><MessageType>1</MessageType><ProcCode>CSNQ</ProcCode><STAN>0220000245250</STAN><LocalTxnDtTime>2203221157</LocalTxnDtTime><DeliveryChannelCtrlID>ATM</DeliveryChannelCtrlID><PName>ACCOUNTNUMBER</PName><PValue>157336</PValue><ActCode>0</ActCode><ActDescription>Success</ActDescription><TotalnoofTrans>1</TotalnoofTrans><Customers><Record><Name>IVAN IVANOV</Name><FirstName>IVAN</FirstName><MiddleName></MiddleName><LastName>IVANOV</LastName><BaseNumber>157336</BaseNumber><Nationality></Nationality><PoBox></PoBox><Address></Address><City></City><Country></Country><Email>example@example.com</Email><CardOnlyCustomer></CardOnlyCustomer><SMSMobile></SMSMobile><SMSLang></SMSLang><SMSNationalID></SMSNationalID><SMSPassportNo></SMSPassportNo><SegmentCode></SegmentCode><SegmentDesc></SegmentDesc><QID>273XXXXXXXX</QID><QIDExpiryDate></QIDExpiryDate><PassportNo>XXXXXXXX</PassportNo><PassportExpiryDate></PassportExpiryDate><CompanyRegNo></CompanyRegNo><CompanyRegNoExpiryDate></CompanyRegNoExpiryDate><LOB></LOB><DOB></DOB><CustTypeFlag></CustTypeFlag></Record></Customers><REFNUM>256557</REFNUM></XML>`,
			),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResponseJ2X(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResponseJ2X() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResponseJ2X() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func BenchmarkResponseJ2X(b *testing.B) {
	tests := []struct {
		name    string
		req     []byte
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"request",
			[]byte(
				`{"RequestInfo":{"requestId":"0220000245250","userId":"256557","basenumber":"157336","chanelId":"ATM","requestTime":"2203221157"},"CustomerDetails":[{"QID":"27340400282","BASENO":"157336","CRNO":null,"PASSPORTNO":"A1810149","MOBILENO":"","EMAILID":"oyange.george73@gmail.com","GUID":"7c7f7a47-f236-ea11-9132-00505685b1c3","FirstName":"GEORGE","LastName":"ONYANGO OYANGE","IsBlacklisted":false,"IsQANationalityWithdrawn":false}]}`,
			),
			[]byte(
				`<XML><MessageType>1</MessageType><ProcCode>CSNQ</ProcCode><STAN>0220000245250</STAN><LocalTxnDtTime>2203221157</LocalTxnDtTime><DeliveryChannelCtrlID>ATM</DeliveryChannelCtrlID><PName>ACCOUNTNUMBER</PName><PValue>157336</PValue><ActCode>0</ActCode><ActDescription>Success</ActDescription><TotalnoofTrans>1</TotalnoofTrans><Customers><Record><Name>GEORGE ONYANGO OYANGE</Name><FirstName>GEORGE</FirstName><MiddleName>ONYANGO</MiddleName><LastName>OYANGE</LastName><BaseNumber>157336</BaseNumber><Nationality></Nationality><PoBox></PoBox><Address></Address><City></City><Country></Country><Email>oyange.george73@gmail.com</Email><CardOnlyCustomer></CardOnlyCustomer><SMSMobile></SMSMobile><SMSLang></SMSLang><SMSNationalID></SMSNationalID><SMSPassportNo></SMSPassportNo><SegmentCode></SegmentCode><SegmentDesc></SegmentDesc><QID>27340400282</QID><QIDExpiryDate></QIDExpiryDate><PassportNo>A1810149</PassportNo><PassportExpiryDate></PassportExpiryDate><CompanyRegNo></CompanyRegNo><CompanyRegNoExpiryDate></CompanyRegNoExpiryDate><LOB></LOB><DOB></DOB><CustTypeFlag></CustTypeFlag></Record></Customers><REFNUM>256557</REFNUM></XML>`,
			),
			false,
		},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := ResponseJ2X(tt.req)
				if err != nil {
					b.Errorf("ResponseJ2X() error = %v", err)
					return
				}
			}
		})
	}
}
