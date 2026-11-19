#!/usr/bin/env bash
# nginx2root (2026-11-18) exposure check.
#
# CVE-2026-123007 / -123633  HTTP/2 stack          (fixed 1.31.9, 1.30.7)
# CVE-2026-121810            HTTP/3 + upstream     (fixed 1.31.9, 1.30.7)
# CVE-2026-123634            resolver + var proxy_pass + mirror
# CVE-2026-113425            mail proxy (proxy on + auth_http)
#
# Usage:
#   ./check-nginx-cves.sh                     # local nginx + local docker images + default hosts
#   ./check-nginx-cves.sh -i my/nginx:tag     # check extra container image(s)
#   ./check-nginx-cves.sh -H example.com:443  # check ALPN/HTTP2 on extra endpoints
#
# Exit code: 0 = nothing vulnerable found, 1 = at least one VULNERABLE item.

set -uo pipefail

FIXED_MAINLINE=1.31.9
FIXED_STABLE=1.30.7

# Version considered patched?
#   >= 1.31.9          (mainline fix)
#   1.30.7 .. 1.30.x  (stable fix)
is_patched() {
    local v=$1 IFS=. a b c
    # shellcheck disable=SC2086
    set -- $v
    a=${1:-0}; b=${2:-0}; c=${3:-0}
    if [ "$a" -gt 1 ]; then return 0; fi
    if [ "$a" -lt 1 ]; then return 1; fi
    if [ "$b" -gt 31 ]; then return 0; fi
    if [ "$b" -eq 31 ]; then [ "$c" -ge 9 ] && return 0; return 1; fi
    if [ "$b" -eq 30 ]; then [ "$c" -ge 7 ] && return 0; return 1; fi
    return 1
}

BAD=0
IMAGES=()
HOSTS=("homin.dev:443" "llm.homin.dev:443")

while getopts ":i:H:" opt; do
    case $opt in
        i) IMAGES+=("$OPTARG") ;;
        H) HOSTS+=("$OPTARG") ;;
        *) echo "usage: $0 [-i image]... [-H host:port]..." >&2; exit 2 ;;
    esac
done

echo "== 1. local nginx binary"
if command -v nginx >/dev/null 2>&1; then
    RAW=$(nginx -v 2>&1)
    VER=$(printf '%s' "$RAW" | sed -n 's|.*nginx/\([0-9.]*\).*|\1|p')
    if [ -z "$VER" ]; then
        echo "   ? could not parse: $RAW"
    elif is_patched "$VER"; then
        echo "   OK        nginx $VER (patched)"
    else
        echo "   VULNERABLE nginx $VER -> upgrade to $FIXED_MAINLINE / $FIXED_STABLE"
        BAD=1
    fi
else
    echo "   -         no nginx binary in PATH"
fi

echo
echo "== 2. container images"
if command -v docker >/dev/null 2>&1; then
    # Auto-discover local images whose /etc/nginx exists, plus any -i arguments.
    if [ ${#IMAGES[@]} -eq 0 ]; then
        while read -r img; do
            [ -z "$img" ] && continue
            if docker run --rm --pull=never --entrypoint sh "$img" -c 'command -v nginx >/dev/null 2>&1' >/dev/null 2>&1; then
                IMAGES+=("$img")
            fi
        done < <(docker images --format '{{.Repository}}:{{.Tag}}' 2>/dev/null \
                 | grep -v '^<none>' | grep -viE '^(nginx|library/nginx):' | head -40)
    fi

    if [ ${#IMAGES[@]} -eq 0 ]; then
        echo "   -         no candidate images checked (pass -i image:tag)"
    fi

    for img in "${IMAGES[@]}"; do
        RAW=$(docker run --rm --pull=never --entrypoint nginx "$img" -v 2>&1)
        VER=$(printf '%s' "$RAW" | sed -n 's|.*nginx version: nginx/\([0-9.]*\).*|\1|p')
        if [ -z "$VER" ]; then
            echo "   skip      $img (no nginx / not pullable)"
            continue
        fi
        # Does the config put HTTP/2 in play? (edge exposure for -123007 / -123633)
        H2=$(docker run --rm --pull=never --entrypoint sh "$img" -c \
             'grep -rqsE "^\s*(http2\s+on;|listen[^;]*http2)" /etc/nginx/ && echo yes || echo no' 2>/dev/null)
        if is_patched "$VER"; then
            echo "   OK        $img -> nginx $VER (http2-in-config: $H2)"
        else
            echo "   VULNERABLE $img -> nginx $VER (http2-in-config: $H2)"
            BAD=1
        fi
    done
else
    echo "   -         docker not available"
fi

echo
echo "== 3. remote endpoints (ALPN h2 offered = HTTP/2 reachable)"
if command -v openssl >/dev/null 2>&1; then
    for hp in "${HOSTS[@]}"; do
        host=${hp%%:*}; port=${hp##*:}
        ALPN=$(openssl s_client -connect "$hp" -servername "$host" -alpn h2 </dev/null 2>/dev/null \
               | grep -i 'ALPN protocol' | head -1 | sed 's/.*: //' | tr -d '[:space:]')
        [ -z "$ALPN" ] && ALPN="none"
        echo "   $hp -> ALPN: $ALPN"
    done
    echo "   note: h2 at the edge is only a problem if an UNPATCHED nginx terminates it."
    echo "         Cloudflare/Traefik front-ends are not affected by these CVEs."
else
    echo "   -         openssl not available"
fi

echo
echo "== 4. config greps to eyeball on any real nginx host"
cat <<'EOF'
   http2 on;                          -> CVE-2026-123007, -123633  (mitigation: http2 off;)
   listen 443 ssl; (with legacy http2)-> same
   quic / http3                       -> CVE-2026-121810          (mitigation: disable HTTP/3)
   resolver + proxy_pass $var         -> CVE-2026-123634          (mitigation: fixed upstream group)
   mail { proxy on; auth_http ... }   -> CVE-2026-113425          (mitigation: keep off untrusted nets)
EOF

echo
if [ "$BAD" -eq 0 ]; then
    echo "RESULT: nothing vulnerable found."
else
    echo "RESULT: vulnerable nginx found above - rebuild on $FIXED_MAINLINE / $FIXED_STABLE and RESTART (reload keeps old workers)."
fi
exit "$BAD"
