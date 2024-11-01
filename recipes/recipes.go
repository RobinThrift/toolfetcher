package recipes

type Recipe struct {
	Name string

	Src Source

	Test []string

	Arch map[string]string
	OS   map[string]string
}

type SourceType string

const (
	SourceTypeGoInstall   SourceType = "goinstall"
	SourceTypeBinDownload SourceType = "bindownload"
)

type Source struct {
	Type        SourceType
	URLTemplate string
	BinPath     string
}
