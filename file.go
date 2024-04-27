package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ThreadCount int
	Parallel    int
	BufferSize  uint8
}

type File struct {
	Url    string
	SaveAs string
	client http.Client
	config Config
	meta   Meta
}

type Meta struct {
	Size      int
	Name      string
	Resumable bool
	Headers   map[string][]string
}

type Progress struct {
	ChunkId    int
	Percentage int
	Completed  bool
	Error      bool
	Message    string
}

func (file File) New(url string, config Config) *File {
	return &File{
		Url:    url,
		client: http.Client{},
		config: config,
	}
}

func (file *File) Meta() *Meta {
	head, error := file.client.Head(file.Url)

	meta := Meta{}

	if error == nil {
		headers := head.Header

		meta.Headers = headers

		size, error := strconv.Atoi(headers.Get("Content-Length"))

		if error != nil {
			fmt.Println("Unable to fetch file meta.")

			return nil
		} else {
			name, _ := file.GuessFileName(meta)

			meta.Size = size
			meta.Name = name
			meta.Resumable = file.IsResumable(meta)

			file.meta = meta

			return &meta
		}
	} else {
		fmt.Println("Unable to fetch file meta.")

		return nil
	}
}

func (file *File) IsResumable(meta Meta) bool {
	_, has := meta.Headers["Accept-Ranges"]

	if has {
		return true
	} else {
		return false
	}
}

func (file *File) GuessFileName(meta Meta) (string, bool) {
	values, has := meta.Headers["Content-Disposition"]

	if has {
		return file.GuessFileNameUsingContentDisposition(values)
	}

	values, has = meta.Headers["Content-Type"]

	if has {
		return file.GuessFileNameUsingContentType(values)
	}

	url, error := url.Parse(file.Url)

	if error == nil {
		return file.GuessFileNameUsingPath(*url)
	}

	return "", false
}

func (file *File) GuessFileNameUsingContentDisposition(values []string) (string, bool) {
	segments := strings.Split(values[0], ";")

	for _, segment := range segments {
		if strings.HasPrefix(segment, "application/") {
			return segment[12:], true
		}
	}
	return "", false
}

func (file *File) GuessFileNameUsingContentType(values []string) (string, bool) {
	segments := strings.Split(values[0], ";")

	for _, segment := range segments {
		if strings.HasPrefix(segment, "application/") {
			return segment[12:], true
		}
	}

	return "", false
}

func (file *File) GuessFileNameUsingPath(url url.URL) (string, bool) {
	path := url.Path

	lastIndexOfDot := strings.LastIndex(path, ".")

	lastIndexOfSlash := strings.LastIndex(path, "/")

	name := "Unknown"

	extension := ".unk"

	if lastIndexOfDot != -1 {
		extension = path[lastIndexOfDot:]

		if lastIndexOfSlash != -1 {
			name = path[lastIndexOfSlash+1 : lastIndexOfDot]
		} else {
			name = "Unknown"
		}
	}

	return fmt.Sprintf("%s.%s", name, extension), true
}

func (file *File) Download(saveAs string, channel chan<- Progress) {
	file.SaveAs = saveAs

	size := file.meta.Size

	chunkSize := size / file.config.ThreadCount

	offset := 0

	for i := 0; i < file.config.ThreadCount; i++ {
		file.downloadChunk(channel, i+1, offset, offset+chunkSize)

		offset += chunkSize
	}
}

func (file *File) downloadChunk(channel chan<- Progress, cID int, offsetFrom int, offsetTo int) {
	response, error := file.client.Get(file.Url)

	if error != nil {
		channel <- Progress{
			ChunkId:    cID,
			Error:      true,
			Completed:  false,
			Percentage: 0,
			Message:    error.Error(),
		}
	} else {
		f, err := os.Create(fmt.Sprintf("%s.chunk_%d", file.SaveAs, cID))

		if err != nil {
			channel <- Progress{
				ChunkId:    cID,
				Error:      true,
				Completed:  false,
				Percentage: 0,
				Message:    "Unable to create file chunk",
			}
		}

		buffer := make([]byte, file.config.BufferSize)

		reader := response.Body

		defer reader.Close()

		defer f.Close()

		_, err = reader.Read(buffer)

		for err != io.EOF {
			f.Write(buffer)

			_, err = reader.Read(buffer)
		}
	}
}
