// Package assets はアプリケーションに埋め込むリソースを提供します
package assets

import _ "embed"

// IconPNG はアプリケーションアイコン（PNG）です。再生成: go run assets/genicon.go
//
//go:embed icon.png
var IconPNG []byte
