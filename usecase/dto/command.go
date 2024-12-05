package dto

import (
	"io"
)

type CommandImageMetadataResult struct {
	Type       string `json:"type"`
	Sha256Hash string `json:"sha256"`
	Size       int    `json:"size"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
}

type FileContentWrapper struct {
	Content io.ReadCloser
	Type    string
	Size    int64
}
