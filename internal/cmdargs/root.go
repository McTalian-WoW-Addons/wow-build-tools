package cmdargs

type RootArgs struct {
	LevelVerbose bool
	LevelDebug   bool
	NoEmoji      bool
	NoColor      bool
	Boring       bool
}

var RootParams *RootArgs = &RootArgs{
	LevelVerbose: false,
	LevelDebug:   false,
	NoEmoji:      false,
	NoColor:      false,
	Boring:       false,
}
