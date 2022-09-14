package models

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestInsertRetrieveMedia(t *testing.T) {
	tuncateMedia()
	SetStoragePath("/tmp/")

	path := fmt.Sprintf("/path/to/media_%d", time.Now().Unix())

	content := "This is the content"
	md := &Media{
		Path:        path,
		FileName:    "file_name.img",
		Description: "description",
		Mime:        "img/png",
		Created:     time.Now(),
		Modified:    time.Now(),
	}

	err := md.StoreContent([]byte(content))
	if err != nil {
		t.Error("error storing content:", err)
	}

	fileContent, err := md.GetContent()
	if err != nil {
		t.Error("error reading file:", err)
	}

	if content != string(fileContent) {
		t.Error("expected:", content, "returned:", string(fileContent))
	}

	newMd := FindMedia(path)
	md.Hash = newMd.Hash
	// TODO: Fix TZs
	md.Created = newMd.Created
	md.Modified = newMd.Modified
	if !reflect.DeepEqual(md, newMd) {
		t.Error("the returned media:", newMd, "is not the same as stored:", md)
	}
}
