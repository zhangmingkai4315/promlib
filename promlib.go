package promlib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

const VersionV1 = "/api/v1"
const QueryAPI = "/query"

const (
	StatusSuccess = "success"
	StatusError = "error"
)

const (
	ResponseTypeMatrix = "matrix"
	ResponseTypeVector = "vector"
	ResponseTypeScalars = "scalars"
	ResponseTypeString = "string"
)

type PromJob struct {
	endpoint string `validate:"required, url"`
	query    string `validate:"required"`
	method   string `validate:"required"`
	api      string `validate:"required"`
	version  string `validate:"required"`
}

func NewPromQueryJob(endpoint string, query string, method string) (*PromJob, error) {
	if method != http.MethodGet && method != http.MethodPost {
		return nil, errors.New("method not allow")
	}
	validate := validator.New()
	job := PromJob{
		endpoint: endpoint,
		query:    query,
		method:   method,
		version:  VersionV1,
		api:      QueryAPI,
	}
	err := validate.Struct(&job)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

type QueryPost struct {
	Query string `json:"query"`
}

type QueryResponseDataResultMetric struct {
	Name     string `json:"__name__"`
	Job      string `json:"job"`
	Instance string `json:"instance"`
}
type QueryResponseVectorResult struct {
	Metric  QueryResponseDataResultMetric `json:"metric"`
	Value   []interface{}                `json:"value"`
}

type QueryResponseMatrixResult struct {
	Metric  QueryResponseDataResultMetric `json:"metric"`
	Values   []interface{}                 `json:"values"`
}

type QueryResponseData struct {
	ResponseType string                    `json:"resultType"`
	Result       []interface{}             `json:"result"`
}

type QueryResponse struct {
	Status string            `json:"status"`
	Data   QueryResponseData `json:"data"`
	Error  string            `json:"error"`
	ErrorType string         `json:"errorType"`
}

type ResponseDataSet struct {
	Metric QueryResponseDataResultMetric
	Data []interface{}
}

func (response *QueryResponse)GetDataSet()[]ResponseDataSet{
	results:=make([]ResponseDataSet,0)
	switch response.Data.ResponseType {
	case ResponseTypeMatrix:
		for _, data := range response.Data.Result{
			result := QueryResponseMatrixResult{}
			err := mapstructure.Decode(data, &result)
			if err != nil{
				continue
			}
			results = append(results, ResponseDataSet{
				Metric: result.Metric,
				Data:   result.Values,
			})
		}
	case ResponseTypeVector:
		for _, data := range response.Data.Result{
			result := QueryResponseVectorResult{}
			err := mapstructure.Decode(data, &result)
			if err != nil{
				continue
			}
			results = append(results, ResponseDataSet{
				Metric: result.Metric,
				Data:   result.Value,
			})
		}
	default:
		for _, data := range response.Data.Result{
			results = append(results, ResponseDataSet{
				Data:   []interface{}{data},
			})
		}
	}
	return results
}

func (job *PromJob) Query() (*QueryResponse,error) {
	var err error
	var resp *http.Response

	if job.method == http.MethodGet {
		url := fmt.Sprintf("%s%s%s?query=%s", job.endpoint, job.version, job.api, job.query)
		resp, err = http.Get(url)
	} else {
		url := fmt.Sprintf("%s%s%s", job.endpoint, job.version, job.api)
		queryPostData := QueryPost{Query: job.query}
		req, _ := json.Marshal(queryPostData)
		resp, err = http.Post(url, "application/json", bytes.NewBuffer(req))
	}

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	queryResponse := QueryResponse{}
	err = json.NewDecoder(resp.Body).Decode(&queryResponse)
	if err != nil{
		return nil, err
	}
	if queryResponse.Status != StatusSuccess{
		return nil, fmt.Errorf("%s: %s",queryResponse.ErrorType, queryResponse.Error)
	}
	return &queryResponse, nil
}


