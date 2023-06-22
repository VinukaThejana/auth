#!/bin/bash

usql "$(awk -F "=" '/DATABASE_URL/ {print $2}' .env | sed 's/.\{7\}$//')" || go install github.com/xo/usql@latest
