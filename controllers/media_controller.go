package controllers

import (
	"net/http"
	"sync"
)

type MediaController struct {
	basePath    string
	repoObjects *sync.Map
}

func GetMediaController() (ds *MediaController) {
	return &MediaController{
		repoObjects: &sync.Map{},
	}
}

func (ds *MediaController) SetBasePath(basePath string) {
	ds.basePath = basePath
}

func (ds *MediaController) GetBasePath() string {
	return ds.basePath
}

func (*MediaController) GetName() string {
	return "Data Storage"
}

func (ds *MediaController) Show(repository, resourceId string) (int, interface{}) {
	if repoObjs, ok := ds.repoObjects.Load(repository); ok {
		if data, ok := repoObjs.(*sync.Map).Load(resourceId); ok {
			return http.StatusOK, data
		}
	}
	return http.StatusNotFound, "Not Found"
}

func (ds *MediaController) New(repository string, body []byte) (int, interface{}) {
	// A possible option is to use the SHA256 of the body as ID in order to
	// avoid duplication, but since we don't have ownership, the destroy
	// may cause lateral issues
	oidStr := "ewlkfnewf"

	// We need to do a lock in all the map since we may endup with two objects
	// initializing the same repository and overriding each other
	repoObjs, _ := ds.repoObjects.LoadOrStore(repository, &sync.Map{})
	repoObjs.(*sync.Map).Store(oidStr, body)

	result := map[string]interface{}{
		"oid":  oidStr,
		"size": len(body),
	}

	return http.StatusCreated, result
}

func (ds *MediaController) Destroy(repository, resourceId string) (int, interface{}) {
	if repoObjs, ok := ds.repoObjects.Load(repository); ok {
		_, loaded := repoObjs.(*sync.Map).LoadAndDelete(resourceId)
		if !loaded {
			return http.StatusNotFound, "Not Found"
		}

		return http.StatusOK, "OK"
	}
	return http.StatusNotFound, "Not Found"
}
