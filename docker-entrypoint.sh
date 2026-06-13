#!/bin/sh
set -e

mkdir -p public/uploads/tus public/files public/exports scorm-packages

if [ "${RUN_MIGRATIONS:-true}" = "true" ]; then
  echo "Running database migrations..."
  ./myapp migrate
fi

echo "Starting application..."
exec ./myapp
