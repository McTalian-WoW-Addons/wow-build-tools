package upload

type UploadArgs struct {
	Input             string
	Label             string
	InterfaceVersions []int
	Changelog         string
	ReleaseType       string
}

var UploadParams = &UploadArgs{}
