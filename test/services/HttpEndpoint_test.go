package test_services

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	cconf "github.com/pip-services3-go/pip-services3-commons-go/config"
	crefer "github.com/pip-services3-go/pip-services3-commons-go/refer"
	"github.com/pip-services3-go/pip-services3-rpc-go/services"
	tdata "github.com/pip-services3-go/pip-services3-rpc-go/test/data"
	tlogic "github.com/pip-services3-go/pip-services3-rpc-go/test/logic"
	"github.com/stretchr/testify/assert"
)

func TestHttpEndpoint(t *testing.T) {

	restConfig := cconf.NewConfigParamsFromTuples(
		"connection.protocol", "http",
		"connection.host", "localhost",
		"connection.port", "3000",
	)

	var endpoint *services.HttpEndpoint
	var service *DummyRestService

	ctrl := tlogic.NewDummyController()
	service = NewDummyRestService()
	service.Configure(cconf.NewConfigParamsFromTuples(
		"base_route",
		"/api/v1",
	))

	endpoint = services.NewHttpEndpoint()
	endpoint.Configure(restConfig)

	references := crefer.NewReferencesFromTuples(
		crefer.NewDescriptor("pip-services-dummies", "controller", "default", "default", "1.0"), ctrl,
		crefer.NewDescriptor("pip-services-dummies", "service", "rest", "default", "1.0"), service,
		crefer.NewDescriptor("pip-services", "endpoint", "http", "default", "1.0"), endpoint,
	)
	service.SetReferences(references)

	err := endpoint.Open("")

	if err != nil {
		assert.Nil(t, err)
	} else {
		defer endpoint.Close("")
		err = service.Open("")
		if err != nil {
			assert.Nil(t, err)
		} else {
			defer service.Close("")
		}
	}

	url := "http://localhost:3000"

	getResponse, getErr := http.Get(url + "/api/v1/dummies")
	assert.Nil(t, getErr)
	resBody, bodyErr := ioutil.ReadAll(getResponse.Body)
	assert.Nil(t, bodyErr)
	var dummies tdata.DummyDataPage
	jsonErr := json.Unmarshal(resBody, &dummies)
	assert.Nil(t, jsonErr)
	assert.NotNil(t, dummies)
	assert.Len(t, dummies.Data, 0)
}
