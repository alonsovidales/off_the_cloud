package models

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"time"
)

type Media struct {
	Path        string
	Hash        string
	FileName    string
	Tags        []string
	Description string
	Mime        string
	Created     time.Time
	Modified    time.Time
}

var storagePath string

func SetStoragePath(path string) {
	storagePath = path
}

func (md *Media) Save() error {
	existingValue := FindMedia(md.Path)
	if existingValue != nil {
		log.Println("Existing value:", existingValue)
	}

	_, err := execStm("INSERT INTO media(path, hash, file_name, description, mime, created, modified) VALUES (?, ?, ?, ?, ?, ?, ?)",
		md.Path,
		md.Hash,
		md.FileName,
		md.Description,
		md.Mime,
		md.Created,
		md.Modified,
	)

	return err
}

func (md *Media) StoreContent(content []byte) error {
	md.Hash = fmt.Sprintf("%x", sha256.Sum256(content))

	err := os.WriteFile(md.getLocalPath(), content, 0644)
	if err != nil {
		panic(fmt.Sprintf("It was not possible to write file on disk: %x", err))
	}

	return md.Save()
}

func (md *Media) GetContent() ([]byte, error) {
	return os.ReadFile(md.getLocalPath())
}

func FindMedia(path string) *Media {
	md := new(Media)
	err := getRow(
		context.Background(),
		"SELECT `path`, `hash`, `file_name`, `description`, `mime`, `created`, `modified` FROM media WHERE `path` = ?",
		path).Scan(
		&md.Path,
		&md.Hash,
		&md.FileName,
		&md.Description,
		&md.Mime,
		&md.Created,
		&md.Modified,
	)

	if err != nil {
		return nil
	}

	return md
}

// tuncateMedia only for testing
func tuncateMedia() {
	execStm("TRUNCATE media")
}

func (md *Media) getLocalPath() string {
	return fmt.Sprintf("%s%s", storagePath, md.Hash)
}
