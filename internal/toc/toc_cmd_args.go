package toc

type TocArgs struct {
	AddonDir string
	Beta     bool
	Ptr      bool
}

var TocParams *TocArgs = &TocArgs{
	AddonDir: ".",
	Beta:     false,
	Ptr:      false,
}

type TocCheckArgs struct {
	IgnoreFiles           []string
	SkipInterfaceCheck    bool
	SkipMissingFilesCheck bool
	SkipNameCheck         bool
}

var TocCheckParams *TocCheckArgs = &TocCheckArgs{
	IgnoreFiles:           []string{},
	SkipInterfaceCheck:    false,
	SkipMissingFilesCheck: false,
	SkipNameCheck:         false,
}
