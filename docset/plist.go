package docset

import (
	"html/template"
	"io"
)

type tplVars struct {
	*Docset
	HasToc bool
}

func renderPlist(w io.Writer, d *Docset) error {
	vars := &tplVars{
		Docset: d, HasToc: true,
	}
	if err := plist.Execute(w, vars); err != nil {
		return err
	}
	return nil
}

// https://github.com/zealdocs/zeal/issues/383#issuecomment-124315538
var plist = template.Must(template.New("").Parse(`
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleIdentifier</key>      <string>{{.Ident}}</string>
	<key>CFBundleName</key>            <string>{{.Name}}</string>
	<key>DocSetPlatformFamily</key>    <string>{{.PlatformFamily}}</string>
	<key>isJavaScriptEnabled</key>     {{if .AllowJs}}<true/>{{else}}<false/>{{end}}
	<key>isDashDocset</key>            <true/>
	<key>dashIndexFilePath</key>       <string>{{.IndexFile}}</sttring>
	<DashDocSetBlocksOnlineResources>  {{if .AllowOnline}}<true/>{{else}}<false/>{{end}}

	{{if .WebSearchKeyword -}}
	<key>DashWebSearchKeyword</key>    <string>{{.WebSearchKeyword}}</string>
	{{end}}

	{{if .HasToc -}}
	<key>DashDocSetFamily</key>        <string>dashtoc</string>
	{{end}}
</dict>
</plist>
`))
