package controllers

import (
	"net/http"
	"net/url"
)

type MediaController struct {
	basePath string
}

func GetMediaController() (ds *MediaController) {
	return &MediaController{}
}

func (ds *MediaController) IsBinary() bool {
	return false
}

func (ds *MediaController) SetBasePath(basePath string) {
	ds.basePath = basePath
}

func (ds *MediaController) GetBasePath() string {
	return ds.basePath
}

func (*MediaController) GetName() string {
	return "Media Controller"
}

func (ds *MediaController) Show(repository, resourceId string) (int, interface{}) {
	return http.StatusNotFound, "Not Found"
}

func (ds *MediaController) Edit(resourceId, extraUrl string, params url.Values, binBody []byte) (int, interface{}) {
	return http.StatusCreated, ""
}

func (ds *MediaController) New(params url.Values, binBody []byte) (int, interface{}) {
	return http.StatusCreated, ""
}

func (ds *MediaController) Destroy(repository, resourceId string) (int, interface{}) {
	return http.StatusOK, "OK"
	return http.StatusNotFound, "Not Found"
}
