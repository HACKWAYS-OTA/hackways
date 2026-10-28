#!/usr/bin/env sh
# Certbot deploy hook for this legacy-domain certificate only.
set -eu
case " ${RENEWED_DOMAINS:-} " in
  *" hackways.me "*)
    docker exec hackways-nginx nginx -t
    docker exec hackways-nginx nginx -s reload
    ;;
esac
