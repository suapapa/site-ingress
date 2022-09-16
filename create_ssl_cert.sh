#!/bin/sh

yes Y | \
certbot certonly --webroot --webroot-path /tmp/letsencrypt \
  -m "ff4500@gmail.com" -d "homin.dev" \
  --agree-tos --non-interactive