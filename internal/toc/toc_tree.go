package toc

type TocTree struct {
	Entries []*TocEntry
}

func (t *TocTree) FlattenEntries() []string {
	var files []string
	var flatten func(entries []*TocEntry)
	flatten = func(entries []*TocEntry) {
		for _, entry := range entries {
			files = append(files, entry.Filepath)
			if len(entry.Entries) > 0 {
				flatten(entry.Entries)
			}
		}
	}
	flatten(t.Entries)
	return files
}
