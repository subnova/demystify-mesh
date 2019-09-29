#!/usr/bin/env sh

set -e

exec su-exec app:app ./control "$@"