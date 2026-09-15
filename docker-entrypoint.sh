#!/bin/sh
set -e

# Railway Volume: one mount (RAILWAY_VOLUME_MOUNT_PATH). Persist uploads + SCORM.
# Source: https://docs.railway.com/guides/volumes
if [ -n "${RAILWAY_VOLUME_MOUNT_PATH:-}" ]; then
  DATA="${RAILWAY_VOLUME_MOUNT_PATH}"
  mkdir -p "$DATA/public/uploads/tus" "$DATA/public/files" "$DATA/public/exports" "$DATA/scorm-packages"
  rm -rf public scorm-packages
  ln -sfn "$DATA/public" public
  ln -sfn "$DATA/scorm-packages" scorm-packages
else
  mkdir -p public/uploads/tus public/files public/exports scorm-packages
fi

if [ "${RUN_MIGRATIONS:-true}" = "true" ]; then
  echo "Running database migrations..."
  ./myapp migrate
fi

echo "Starting application..."
exec ./myapp
