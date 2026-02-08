#!/bin/bash
echo "Setup [tessl]: Initializing Tessl environment"

# login_method: native = tessl init, env = skip (no env-based auth), auto = native
if [ "$ADDT_EXT_AUTO_LOGIN" = "true" ]; then
    method="${ADDT_EXT_LOGIN_METHOD:-auto}"

    if [ "$method" = "native" ] || [ "$method" = "auto" ]; then
        echo "Setup [tessl]: Auto-initializing Tessl"
        tessl init
    fi
fi
