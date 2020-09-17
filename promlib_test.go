package promlib

import (
	"log"
	"net/http"
	"testing"
)

func TestNewPromQueryJob(t *testing.T) {
	endpoint := "http://127.0.0.1:9090"
	job,err := NewPromQueryJob(endpoint,"up",http.MethodGet)
	if err != nil{
		t.Errorf("expect create job success, but got err %s",err.Error())
	}

	resp,err := job.Query()
	if err != nil{
		t.Errorf("expect query job success, but got err %s",err.Error())
	}
	dataSet := resp.GetDataSet()
	if len(dataSet) == 0{
		t.Errorf("expect got multi result but got none")
	}
	log.Printf("%+v",dataSet)
}