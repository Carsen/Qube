#!/bin/sh

[ -n "${PUID}" ] && usermod -u "${PUID}" bitcaskd
[ -n "${PGID}" ] && groupmod -g "${PGID}" bitcaskd

printf "Configuring bitcaskd...\n"
[ -z "${DATA}" ] && DATA="/data"
export DATA

printf "Switching UID=%s and GID=%s\n" "${PUID}" "${PGID}"
exec su-exec bitcaskd:bitcaskd "$@"
