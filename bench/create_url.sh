#!/bin/bash

URL="http://localhost:8080/api/url"
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Impla3l1bGxsQGdtYWlsLmNvbSIsInVzZXJfaWQiOjEsImV4cCI6MTc0OTUxNDY2NiwiaWF0IjoxNzQ5NDI4MjY2fQ.JKqxp6fgvo2GE8C4byVYBNxtJOivxPcPW0DOXx0_GT8"

generate_payload() {
  RAND_CODE=$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 6 | head -n 1)
  echo '{
    "original_url": "https://www.google.com/search?q=hello&sca_esv=433326fac3a19cfb&sxsrf=AE3TifMxLQR4954Vh8B-56MCcQkhWiLntg%3A1749426555642&source=hp&ei=eyFGaLy2JaqOvr0P6PisyAU&iflsig=AOw8s4IAAAAAaEYvi8MKOF9a3K8kRmDsYJOjIwXvSbAc&ved=0ahUKEwj836HjgeONAxUqh68BHWg8C1kQ4dUDCBc&uact=5&oq=hello&gs_lp=Egdnd3Mtd2l6IgVoZWxsbzIFEC4YgAQyBRAAGIAEMgUQABiABDILEC4YgAQYxwEYrwEyCxAuGIAEGNEDGMcBMgUQLhiABDIFEAAYgAQyBRAAGIAEMgUQABiABDIFEAAYgARIopABUPKIAVjGjQFwA3gAkAEAmAGPAqAB8BKqAQUwLjYuN7gBA8gBAPgBAZgCEKAClROoAgrCAgcQIxgnGOoCwgINECMY8AUYJxjJAhjqAsICChAjGPAFGCcY6gLCAgoQIxiABBgnGIoFwgIEECMYJ5gDBfEFEN6yJiAPwJXxBY0933eWKxbi8QVzB-t483yy1_EF0tS2Arhl-ZjxBS7H0xawemq48QU4Rk671C_OcPEFXYngy8N713jxBYhfW41CBJ7H8QXqg2uHIeGe25IHBTMuNi43oAfbvQGyBwUwLjYuN7gHjRPCBwYwLjE1LjHIBxw&sclient=gws-wiz",
    "custom_code": "'"$RAND_CODE"'",
    "duration": 1,
    "user_id": 1
  }'
}

export URL TOKEN

send_request() {
  curl -s -X POST "$URL" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "$(generate_payload)"
}

export -f send_request generate_payload

# 100并发，执行200次
seq 1 20000 | xargs -n1 -P200 bash -c 'send_request'
