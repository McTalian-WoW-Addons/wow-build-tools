package cmdimpl

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
