package services

import (
	"net/http"
	"time"

	cconv "github.com/pip-services3-go/pip-services3-commons-go/convert"
	crefer "github.com/pip-services3-go/pip-services3-commons-go/refer"
	cinfo "github.com/pip-services3-go/pip-services3-components-go/info"
)

type StatusOperations struct {
	RestOperations
	startTime   time.Time
	references2 crefer.IReferences
	contextInfo *cinfo.ContextInfo
}

func NewStatusOperations() *StatusOperations {
	//super();
	so := StatusOperations{}
	so.startTime = time.Now()
	so.DependencyResolver.Put("context-info", crefer.NewDescriptor("pip-services", "context-info", "default", "*", "1.0"))
	return &so
}

/*
	Sets references to dependent components.

	@param references 	references to locate the component dependencies.
*/
func (c *StatusOperations) SetReferences(references crefer.IReferences) {
	c.references2 = references
	c.RestOperations.SetReferences(references)

	depRes := c.DependencyResolver.GetOneOptional("context-info")
	if depRes != nil {
		c.contextInfo = depRes.(*cinfo.ContextInfo)
	}
}

func (c *StatusOperations) GetStatusOperation() func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		c.Status(res, req)
	}
}

/*
   Handles status requests

   @param req   an HTTP request
   @param res   an HTTP response
*/
func (c *StatusOperations) Status(res http.ResponseWriter, req *http.Request) {

	id := ""
	if c.contextInfo != nil {
		id = c.contextInfo.ContextId
	}

	name := "Unknown"
	if c.contextInfo != nil {
		name = c.contextInfo.Name
	}

	description := ""
	if c.contextInfo != nil {
		description = c.contextInfo.Description
	}

	uptime := time.Now().Sub(c.startTime)

	properties := make(map[string]string)
	if c.contextInfo != nil {
		properties = c.contextInfo.Properties
	}

	var components []string
	if c.references2 != nil {
		for _, locator := range c.references2.GetAllLocators() {
			components = append(components, cconv.StringConverter.ToString(locator))
		}
	}

	status := make(map[string]interface{})

	status["id"] = id
	status["name"] = name
	status["description"] = description
	status["start_time"] = cconv.StringConverter.ToString(c.startTime)
	status["current_time"] = cconv.StringConverter.ToString(time.Now())
	status["uptime"] = uptime
	status["properties"] = properties
	status["components"] = components

	c.SendResult(res, req, status, nil)
}
