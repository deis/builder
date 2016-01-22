package gitreceive

import (
	"text/template"
)

var (
	dockerBuilderNoCredsTpl = template.Must(template.ParseFiles(dockerBuilderNoCredsTplFile))
	dockerBuilderTpl        = template.Must(template.ParseFiles(dockerBuilderTplFile))
	slugBuilderNoCredsTpl   = template.Must(template.ParseFiles(slugBuilderNoCredsTplFile))
	slugBuilderTpl          = template.Must(template.ParseFiles(slugBuilderTplFile))
)

type dockerBuilderTplData struct {
	Name      string
	TarURL    string
	ImageName string
}

type slugBuilderTplData struct {
	Name         string
	TarURL       string
	PutURL       string
	BuildPackURL string
}
