package main

import "net/http"

func main() {
	responder := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"RequestInfo":{"requestId":"0220000245250","userId":"256557","basenumber":"157336","chanelId":"ATM","requestTime":"2203221157"},"CustomerDetails":[{"QID":"2734XXXXXXX","BASENO":"157336","CRNO":null,"PASSPORTNO":"XXXXXX","MOBILENO":"","EMAILID":"XXXXXX@example.com","GUID":"7c7f7a47-f236-ea11-9132-00505685b1c3","FirstName":"XXXXXX","LastName":"XXXXXX","IsBlacklisted":false,"IsQANationalityWithdrawn":false}]}`))
	})

	http.ListenAndServe(":3030", responder)
}
