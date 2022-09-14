package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"testing"
	"time"
)

const (
	testPort = 8888
)

type TestRequest struct {
	method   string
	path     string
	response string
}

var actionsToTest = map[string]TestRequest{
	"show":    TestRequest{"GET", "repo/id", "OK Show"},
	"new":     TestRequest{"PUT", "", "OK New"},
	"destroy": TestRequest{"DELETE", "repo/super_weird/fancy/id", "OK Destroy"},
}

type MockRestHandler struct {
	basePath string

	times map[string]int
}

func GetMockHandler() *MockRestHandler {
	return &MockRestHandler{
		times: make(map[string]int),
	}
}

func (mk *MockRestHandler) GetName() (resourceName string) {
	return "Mock"
}

func (mk *MockRestHandler) GetBasePath() (path string) {
	return mk.basePath
}

func (mk *MockRestHandler) SetBasePath(path string) {
	mk.basePath = path
}

func (mk *MockRestHandler) Show(repository, resourceId string) (int, interface{}) {
	return mk.mockResponse("show")
}

func (mk *MockRestHandler) New(repository string, body []byte) (int, interface{}) {
	return mk.mockResponse("new")
}

func (mk *MockRestHandler) Destroy(repository, resourceId string) (int, interface{}) {
	return mk.mockResponse("destroy")
}

func (mk *MockRestHandler) mockResponse(method string) (int, interface{}) {
	if _, ok := mk.times[method]; ok {
		mk.times[method] += 1
	} else {
		mk.times[method] = 1
	}

	return 200, actionsToTest[method].response
}

func TestStartAndShutDown(t *testing.T) {
	httpApi := GetAPI(testPort)
	httpApi.AddController("/", GetMockHandler())

	go func() {
		err := httpApi.Start()
		if err != nil {
			t.Fatalf("error while starting HTTP API server, err: %v", err)
		}
	}()

	// Just wait until the server is ready to receive connections
	_, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", testPort), time.Second)
	if err != nil {
		t.Fatalf("The HTTP server is still not listening: %v", err)
	}

	httpApi.Shutdown()

	// Now the server should be down
	_, err = net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", testPort), time.Second)
	if err == nil {
		t.Fatalf("The HTTP server is still listening after the shutdown")
	}
}

func BenchmarkRestActions(b *testing.B) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	// We have to disable the logging in order to be able to see the results for
	// the benchmarking, we should wrap the logger
	log.SetOutput(ioutil.Discard)

	httpApi := GetAPI(testPort)
	basePath := "/base_path/"
	mockHandler := GetMockHandler()
	httpApi.AddController(basePath, mockHandler)

	go func() {
		err := httpApi.Start()
		if err != nil {
			b.Fatalf("error while starting HTTP API server, err: %v", err)
		}
	}()

	addr := fmt.Sprintf("localhost:%d", testPort)

	// Just wait until the server is ready to receive connections
	_, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		b.Fatalf("The HTTP server is still not listening: %v", err)
	}

	// We are going to send some requests per action and check with the mock if
	// we are htting the right action and the response is the expected
	for action, resource := range actionsToTest {
		b.Run(fmt.Sprintf("Action: %s", action), func(b *testing.B) {
			mockHandler.times = make(map[string]int)
			client := new(http.Client)
			req, _ := http.NewRequest(resource.method, fmt.Sprintf("http://%s%s%s", addr, basePath, resource.path), nil)

			for i := 0; i < b.N; i++ {
				resp, err := client.Do(req)
				if err != nil {
					b.Errorf("error while readig index, err: %v", err)
				} else {
					defer resp.Body.Close()
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						b.Errorf("error while reading response body, err: %v", err)
					}

					if string(body) != resource.response {
						b.Errorf("response for index method expected: %s but got: %s", resource.response, string(body))
					}
				}
			}
			if mockHandler.times[action] != b.N {
				b.Errorf("%s action invoked %d times, expected %d times", action, mockHandler.times[action], b.N)
			}
		})
	}

	httpApi.Shutdown()

	// Now the server should be down
	_, err = net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", testPort), time.Second)
	if err == nil {
		b.Fatalf("The HTTP server is still listening after the shutdown")
	}
}

func ExampleGetAPI() {
	// Prepare the API to listen in the local 8080 port
	httpApi := GetAPI(8080)

	// Bind as many handlers as necessary to the corresponding base paths
	httpApi.AddController("/base_path/", GetMockHandler())

	// In order to start the API use a Go routine to don't block the execution
	go func() {
		err := httpApi.Start()
		if err != nil {
			log.Fatalf("error while starting HTTP API server, err: %v", err)
		}
	}()

	// We can wait here for a signal in order to stop the server
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	sig := <-c
	log.Printf("Signal received: %v, exiting...", sig)
	// Stop the server gracefully
	httpApi.Shutdown()
	os.Exit(0)
}
