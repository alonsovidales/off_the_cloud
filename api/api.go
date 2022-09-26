package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
)

// Every Controller has to implement this interface providing the methods in
// it. If one of the Actions is not supported, use the *NotSupported structs in
// order to define this methods with a generic response for each one.
type Controller interface {
	// GetName has to return the human friendly name of the controller in order
	// to be used into logs, metrics and so.
	GetName() (resourceName string)
	// SetBasePath will receive and store the base URI path for this controller.
	SetBasePath(path string)
	// GetBasePath has to return the base path set in SetBasePath.
	GetBasePath() (path string)

	IsBinary() bool

	// Show mapped from the HTTP GET method when a resource ID is specieied
	// after the base URI, ex: /base_path/resouce_id
	// this action has to find and return the resource if any for that
	// resourceId in the data returned parameter and a HTTP code in the
	// httpCode returned parameter.
	Show(resourceId, extraUrl string) (httpCode int, data interface{})
	// New mapped from the HTTP POST method, this action has to create a new
	// resource from the provided values returning it in the data returned
	// parameter and a HTTP code in the httpCode returned parameter.
	New(params url.Values, binBody []byte) (httpCode int, data interface{})
	// Destroy will seek and delete an element with the provided resourceId,
	// it has to return an HTTP code in the returned httpCode parameter and
	// some data in the data returned aprameter.
	Destroy(resourceId, extraUrl string) (httpCode int, data interface{})
	Edit(resourceId, extraUrl string, params url.Values, binBody []byte) (httpCode int, data interface{})
}

type (
	// A ShowNotSupported will be used by the controllers in order to compose
	// the Show method when no supported.
	ShowNotSupported struct{}
	// A NewNotSupported will be used by the controllers in order to compose
	// the New method when no supported.
	NewNotSupported struct{}
	// A DestroyNotSupported will be used by the controllers in order to compose
	// the Destroy method when no supported.
	DestroyNotSupported struct{}
	EditNotSupported    struct{}
)

// Show method used to define the Show action for controllers that don't
// support this.
func (ShowNotSupported) Show(resourceId, extraUrl string) (int, interface{}) {
	return 405, "Show method not implemented by this resource"
}

// New method used to define the New action for controllers that don't
// support this.
func (NewNotSupported) New(body []byte) (int, interface{}) {
	return 405, "New method not implemented by this resource"
}

// Index method used to define the Index action for controllers that don't
// support this.
func (DestroyNotSupported) Destroy(resourceId, extraUrl string) (int, interface{}) {
	return 405, "Destroy method not implemented by this resource"
}

// Edit method used to define the Index action for controllers that don't
// support this.
func (EditNotSupported) Edit(resourceId, extraUrl string) (int, interface{}) {
	return 405, "Edit method not implemented by this resource"
}

// A API is the definition of a full REST HTTP API that will route the
// incomming requests to the corresponding controllers and actions in the
// controllers.
type API struct {
	httpSrv *http.ServeMux
	port    int
	server  *http.Server
}

// GetAPI builds and returns the an REST HTTP API that will listen in the
// specified port.
func GetAPI(port int) *API {
	server := http.NewServeMux()
	return &API{
		port:    port,
		httpSrv: server,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: server,
		},
	}
}

// requestHandler request wrapper that distributes all the incomming request to
// a resource to its corresponding action, this method provides some perfomance
// insights as well.
func (api *API) requestHandler(resource Controller) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		var response interface{}
		var code int

		initMilli := time.Now().Nanosecond()

		urlParts := strings.SplitN(request.URL.Path, resource.GetBasePath(), 2)
		repoObject := strings.SplitN(urlParts[len(urlParts)-1], "/", 2)
		resourceId := repoObject[0]
		var extraUrl string
		if len(repoObject) == 2 {
			extraUrl = repoObject[1]
		}

		var params url.Values
		var body []byte
		var err error
		if resource.IsBinary() {
			body, err = ioutil.ReadAll(request.Body)
		} else {
			err = request.ParseForm()
			params = request.Form
		}
		if err != nil {
			http.Error(rw, "Error while reading request body", 422)
		}
		switch request.Method {
		case GET:
			code, response = resource.Show(resourceId, extraUrl)
		case PUT:
			code, response = resource.New(params, body)
		case POST:
			code, response = resource.Edit(resourceId, extraUrl, params, body)
		case DELETE:
			code, response = resource.Destroy(resourceId, extraUrl)
		default:
			http.Error(rw, "Method not supported", 405)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(code)
		returnType := reflect.TypeOf(response).Kind()
		if returnType == reflect.Map {
			content, err := json.Marshal(response)
			if err != nil {
				log.Printf("[ERROR] The returned data by the controller can't be mashalled as JSON: %v", response)
				http.Error(rw, "Internal server error", 500)
			}
			rw.Write(content)
		} else {
			if returnType == reflect.String {
				io.WriteString(rw, response.(string))
			} else {
				rw.Write(response.([]uint8))
			}
		}

		log.Printf("[%s] - %s - %s: %v ms", request.Method, request.URL.Path, resource.GetName(), (time.Now().Nanosecond()-initMilli)/1000000)
	}
}

// AddController binds a base path to a controller, all the requests with
// an URI matching that path will be enrouted to the correspoding controller
// and action based on the HTTP method
func (api *API) AddController(basePath string, resource Controller) {
	log.Printf("Added controller: %s - %s", basePath, resource.GetName())
	resource.SetBasePath(basePath)
	api.httpSrv.HandleFunc(basePath, api.requestHandler(resource))
}

// Start starts the HTTP server listeining in the specified port
func (api *API) Start() (err error) {
	log.Printf("Starting server in port: %d ...", api.port)
	err = api.server.ListenAndServe()
	if err == http.ErrServerClosed {
		// This is a graceful shutdown
		return nil
	}
	return err
}

// Stop stops the HTTP server for the API
func (api *API) Shutdown() (err error) {
	log.Printf("Shutting down...")
	return api.server.Close()
}
