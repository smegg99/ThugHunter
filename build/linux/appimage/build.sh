#!/usr/bin/env bash
set -euo pipefail
set -x

# Required env vars:
: "${APP_NAME:?set APP_NAME (e.g. thughunter)}"
: "${APP_BINARY:?set APP_BINARY (path to your built binary)}"
: "${ICON_PATH:?set ICON_PATH (path to PNG icon)}"
: "${DESKTOP_FILE:?set DESKTOP_FILE (path to .desktop file)}"

APP_DIR="${APP_NAME}.AppDir"
OUT_APPIMAGE="${APP_NAME}.AppImage"

# Clean previous outputs
rm -rf "${APP_DIR}" ./*.AppImage linuxdeploy-*.AppImage linuxdeploy-plugin-gtk.sh || true

mkdir -p "${APP_DIR}/usr/bin"

install -m 0755 "${APP_BINARY}" "${APP_DIR}/usr/bin/${APP_NAME}"

# Desktop + icon (linuxdeploy expects these at AppDir root)
install -m 0644 "${ICON_PATH}"    "${APP_DIR}/${APP_NAME}.png"
install -m 0644 "${DESKTOP_FILE}" "${APP_DIR}/${APP_NAME}.desktop"

# IMPORTANT:
# linuxdeploy-plugin-gtk's AppRun hook sets GTK_DATA_PREFIX="$APPDIR".
# But linuxdeploy uses the standard AppDir layout with data in $APPDIR/usr/share.
# These symlinks make $APPDIR/share etc. exist, so GTK_DATA_PREFIX works.
ln -sf usr/share "${APP_DIR}/share"
ln -sf usr/lib   "${APP_DIR}/lib"
ln -sf usr/bin   "${APP_DIR}/bin"
ln -sf usr/lib   "${APP_DIR}/lib64"  # some systems probe lib64

ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  LINUXDEPLOY_APPIMAGE="linuxdeploy-x86_64.AppImage" ;;
  aarch64) LINUXDEPLOY_APPIMAGE="linuxdeploy-aarch64.AppImage" ;;
  *)
    echo "Unsupported arch: $ARCH" >&2
    exit 1
    ;;
esac

wget -q -4 -N "https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/${LINUXDEPLOY_APPIMAGE}"
chmod +x "${LINUXDEPLOY_APPIMAGE}"

wget -q -4 -N \
  "https://raw.githubusercontent.com/linuxdeploy/linuxdeploy-plugin-gtk/master/linuxdeploy-plugin-gtk.sh" \
  -O linuxdeploy-plugin-gtk.sh
chmod +x linuxdeploy-plugin-gtk.sh

# Make sure linuxdeploy can discover the plugin (it looks in PATH/cwd)
export PATH="$PWD:$PATH"

export DEPLOY_GTK_VERSION=3
export NO_STRIP=1

./"${LINUXDEPLOY_APPIMAGE}" \
  --appdir "${APP_DIR}" \
  --plugin gtk

HOOK="${APP_DIR}/apprun-hooks/linuxdeploy-plugin-gtk.sh"
if [[ -f "${HOOK}" ]]; then
  if ! grep -q 'GTK_EXE_PREFIX' "${HOOK}"; then
    sed -i $'s|export GTK_DATA_PREFIX="$APPDIR"|export GTK_DATA_PREFIX="$APPDIR/usr"\nexport GTK_EXE_PREFIX="$APPDIR/usr"|' "${HOOK}"
  else
    sed -i 's|export GTK_DATA_PREFIX="$APPDIR"|export GTK_DATA_PREFIX="$APPDIR/usr"|' "${HOOK}"
  fi
else
  echo "ERROR: Expected GTK hook not found at: ${HOOK}" >&2
  exit 1
fi

./"${LINUXDEPLOY_APPIMAGE}" \
  --appdir "${APP_DIR}" \
  --output appimage

# Rename output deterministically
GEN="$(ls -1 "${APP_NAME}"*.AppImage | head -n 1 || true)"
if [[ -z "${GEN}" ]]; then
  echo "ERROR: linuxdeploy did not produce an AppImage" >&2
  exit 1
fi

mv -f "${GEN}" "${OUT_APPIMAGE}"
echo "Built: ${OUT_APPIMAGE}"