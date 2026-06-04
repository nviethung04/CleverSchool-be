#!/bin/sh
set -e

if [ "${RUN_MIGRATIONS:-true}" = "true" ]; then
  echo "Running database migrations..."
  ./myapp migrate
fi

echo "Starting application..."
exec ./myapp
