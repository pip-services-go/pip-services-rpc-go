package connect

import (
	"net/url"
	"strconv"

	cconf "github.com/pip-services3-go/pip-services3-commons-go/config"
	cerr "github.com/pip-services3-go/pip-services3-commons-go/errors"
	crefer "github.com/pip-services3-go/pip-services3-commons-go/refer"
	cauth "github.com/pip-services3-go/pip-services3-components-go/auth"
	ccon "github.com/pip-services3-go/pip-services3-components-go/connect"
)

/*
Helper class to retrieve connections for HTTP-based services abd clients.

In addition to regular functions of ConnectionResolver is able to parse http:// URIs
and validate connection parameters before returning them.

Configuration parameters:

- connection:
  - discovery_key:               (optional) a key to retrieve the connection from [[https://rawgit.com/pip-services-node/pip-services3-components-node/master/doc/api/interfaces/connect.idiscovery.html IDiscovery]]
  - ...                          other connection parameters

- connections:                   alternative to connection
  - [connection params 1]:       first connection parameters
  -  ...
  - [connection params N]:       Nth connection parameters
  -  ...

 References:

- \*:discovery:\*:\*:1.0            (optional) [[https://rawgit.com/pip-services-node/pip-services3-components-node/master/doc/api/interfaces/connect.idiscovery.html IDiscovery]] services

@see [[https://rawgit.com/pip-services-node/pip-services3-components-node/master/doc/api/classes/connect.connectionparams.html ConnectionParams]]
@see [[https://rawgit.com/pip-services-node/pip-services3-components-node/master/doc/api/classes/connect.connectionresolver.html ConnectionResolver]]

Example:

    let config = ConfigParams.fromTuples(
         "connection.host", "10.1.1.100",
         "connection.port", 8080
    );

    let connectionResolver = new HttpConnectionResolver();
    connectionResolver.configure(config);
    connectionResolver.setReferences(references);

    connectionResolver.resolve("123", (err, connection) => {
	// Now use connection...
    });
*/
// implements IReferenceable, IConfigurable
type HttpConnectionResolver struct {
	//The base connection resolver.
	ConnectionResolver ccon.ConnectionResolver
	//The base credential resolver.
	CredentialResolver cauth.CredentialResolver
}

func NewHttpConnectionResolver() *HttpConnectionResolver {
	return &HttpConnectionResolver{*ccon.NewEmptyConnectionResolver(), *cauth.NewEmptyCredentialResolver()}
}

/*
   Configures component by passing configuration parameters.
   @param config    configuration parameters to be set.
*/
func (c *HttpConnectionResolver) Configure(config *cconf.ConfigParams) {
	c.ConnectionResolver.Configure(config)
	c.CredentialResolver.Configure(config)
}

/*
	Sets references to dependent components.

	@param references 	references to locate the component dependencies.
*/
func (c *HttpConnectionResolver) SetReferences(references crefer.IReferences) {
	c.ConnectionResolver.SetReferences(references)
	c.CredentialResolver.SetReferences(references)
}

func (c *HttpConnectionResolver) validateConnection(correlationId string, connection *ccon.ConnectionParams, credential *cauth.CredentialParams) error {
	if connection == nil {
		return cerr.NewConfigError(correlationId, "NO_CONNECTION", "HTTP connection is not set")
	}
	uri := connection.Uri()
	if uri != "" {
		return nil
	}

	protocol := connection.Protocol() //"http"
	if "http" != protocol && "https" != protocol {
		return cerr.NewConfigError(correlationId, "WRONG_PROTOCOL", "Protocol is not supported by REST connection").WithDetails("protocol", protocol)
	}
	host := connection.Host()
	if host == "" {
		return cerr.NewConfigError(correlationId, "NO_HOST", "Connection host is not set")
	}
	port := connection.Port()
	if port == 0 {
		return cerr.NewConfigError(correlationId, "NO_PORT", "Connection port is not set")
	}
	// Check HTTPS credentials
	if protocol == "https" {
		// Check for credential
		if credential == nil {
			return cerr.NewConfigError(correlationId, "NO_CREDENTIAL", "SSL certificates are not configured for HTTPS protocol")
		} else {
			if credential.GetAsNullableString("ssl_key_file") == nil {
				return cerr.NewConfigError(
					correlationId, "NO_SSL_KEY_FILE", "SSL key file is not configured in credentials")
			} else if credential.GetAsNullableString("ssl_crt_file") == nil {
				return cerr.NewConfigError(
					correlationId, "NO_SSL_CRT_FILE", "SSL crt file is not configured in credentials")
			}
		}
	}

	return nil
}

func (c *HttpConnectionResolver) updateConnection(connection *ccon.ConnectionParams) {
	if connection == nil {
		return
	}

	uri := connection.Uri()

	if uri == "" {
		protocol := connection.Protocol() // "http"
		host := connection.Host()
		port := connection.Port()

		uri := protocol + "://" + host
		if port != 0 {
			uri += ":" + strconv.Itoa(port)
		}
		connection.SetUri(uri)
	} else {
		address, _ := url.Parse(uri)
		//protocol := ("" + address.protocol).replace(":", "")
		protocol := address.Scheme

		connection.SetProtocol(protocol)
		connection.SetHost(address.Hostname())
		port, _ := strconv.Atoi(address.Port())
		connection.SetPort(port)
	}
}

/*
Resolves a single component connection. If connections are configured to be retrieved
from Discovery service it finds a IDiscovery and resolves the connection there.

@param correlationId     (optional) transaction id to trace execution through call chain.
@param callback 			callback function that receives resolved connection or error.
*/
func (c *HttpConnectionResolver) Resolve(correlationId string) (connection *ccon.ConnectionParams, credential *cauth.CredentialParams, err error) {

	connection, err = c.ConnectionResolver.Resolve(correlationId)
	if err != nil {
		return nil, nil, err
	}

	credential, err = c.CredentialResolver.Lookup(correlationId)
	if err == nil {
		err = c.validateConnection(correlationId, connection, credential)
	}
	if err == nil && connection != nil {
		c.updateConnection(connection)
	}

	return connection, credential, err
}

/*
Resolves all component connection. If connections are configured to be retrieved
from Discovery service it finds a IDiscovery and resolves the connection there.

@param correlationId     (optional) transaction id to trace execution through call chain.
@param callback 			callback function that receives resolved connections or error.
*/
func (c *HttpConnectionResolver) ResolveAll(correlationId string) (connections []*ccon.ConnectionParams, credential *cauth.CredentialParams, err error) {

	connections, err = c.ConnectionResolver.ResolveAll(correlationId)
	if err != nil {
		return nil, nil, err
	}

	credential, err = c.CredentialResolver.Lookup(correlationId)
	if connections == nil {
		connections = make([]*ccon.ConnectionParams, 0)
	}

	for _, connection := range connections {
		if err == nil {
			err = c.validateConnection(correlationId, connection, credential)
		}
		if err == nil && connection != nil {
			c.updateConnection(connection)
		}
	}
	return connections, credential, err
}

/*
Registers the given connection in all referenced discovery services.
c method can be used for dynamic service discovery.

@param correlationId     (optional) transaction id to trace execution through call chain.
@param connection        a connection to register.
@param callback          callback function that receives registered connection or error.
*/
func (c *HttpConnectionResolver) Register(correlationId string) error {

	connection, err := c.ConnectionResolver.Resolve(correlationId)
	if err != nil {
		return err
	}

	credential, err := c.CredentialResolver.Lookup(correlationId)
	// Validate connection
	if err == nil {
		err = c.validateConnection(correlationId, connection, credential)
	}
	if err == nil {
		return c.ConnectionResolver.Register(correlationId, connection)
	} else {
		return err
	}
}
