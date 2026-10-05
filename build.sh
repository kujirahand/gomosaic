#!/bin/bash

# ゴモザイク ビルドスクリプト

set -e

APP_NAME="gomosaic"
VERSION="1.0.1"

echo "🔨 ゴモザイク v${VERSION} をビルドします..."

# ビルドディレクトリを作成
mkdir -p dist

# 現在のプラットフォーム用にビルド
echo "📦 現在のプラットフォーム用にビルド中..."
if [[ "$OSTYPE" == "darwin"* ]]; then
    APP_BUNDLE="dist/${APP_NAME}.app"
    # 古い成果物を削除し、バイナリを.app内に直接出力する（単体のgomosaicは作らない）
    rm -rf "${APP_BUNDLE}" "dist/${APP_NAME}"
    mkdir -p "${APP_BUNDLE}/Contents/MacOS" "${APP_BUNDLE}/Contents/Resources"
    go build -o "${APP_BUNDLE}/Contents/MacOS/${APP_NAME}" ./cmd/gomosaic/

    # アイコン(.icns)を生成してバンドルに入れる
    ICONSET="$(mktemp -d)/AppIcon.iconset"
    mkdir -p "${ICONSET}"
    for sz in 16 32 128 256 512; do
        sips -z ${sz} ${sz} assets/icon.png --out "${ICONSET}/icon_${sz}x${sz}.png" >/dev/null
        sips -z $((sz*2)) $((sz*2)) assets/icon.png --out "${ICONSET}/icon_${sz}x${sz}@2x.png" >/dev/null
    done
    iconutil -c icns "${ICONSET}" -o "${APP_BUNDLE}/Contents/Resources/AppIcon.icns"

    cat > "${APP_BUNDLE}/Contents/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleDevelopmentRegion</key>
    <string>ja</string>
    <key>CFBundleExecutable</key>
    <string>${APP_NAME}</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
    <key>CFBundleIdentifier</key>
    <string>com.github.gomosaic</string>
    <key>CFBundleName</key>
    <string>ゴモザイク</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>${VERSION}</string>
    <key>CFBundleVersion</key>
    <string>1</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.13</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSPrincipalClass</key>
    <string>NSApplication</string>
</dict>
</plist>
EOF

    # ad-hoc署名（.app単体で起動できるようにする）
    codesign --force --deep --sign - "${APP_BUNDLE}" 2>/dev/null || true

    echo "📊 ファイルサイズ:"
    ls -lh "${APP_BUNDLE}/Contents/MacOS/${APP_NAME}"
    echo ""
    echo "🚀 実行方法:"
    echo "   open dist/${APP_NAME}.app"
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    go build -o "dist/${APP_NAME}" ./cmd/gomosaic/
    echo "📊 ファイルサイズ:"
    ls -lh "dist/${APP_NAME}"
    echo ""
    echo "🚀 実行方法:"
    echo "   ./dist/${APP_NAME}"
elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
    go build -o "dist/${APP_NAME}.exe" ./cmd/gomosaic/
    echo "📊 ファイルサイズ:"
    ls -lh "dist/${APP_NAME}.exe"
    echo ""
    echo "🚀 実行方法:"
    echo "   dist\\${APP_NAME}.exe"
fi

echo ""
echo "✨ ビルドが完了しました！"
