#!/usr/bin/env bash
# Generates the CA used by si-inner to sign certificates for intercepted HTTPS requests.
#
# Usage: ./generate-ca.sh [output-dir] [common-name]
#
# The private key is written with mode 0600 and must never be committed to version control.
# Distribute ONLY the certificate (myCA.pem) to the inner clients' trust stores.
set -euo pipefail

OUT="${1:-/safe-import/ca}"
CN="${2:-safe-import inner CA ($(hostname))}"
DAYS="${CA_DAYS:-1825}"

mkdir -p "$OUT"
chmod 700 "$OUT"

if [ -e "$OUT/myCA.key" ]; then
    echo "ERROR: $OUT/myCA.key already exists, remove it first if you want to replace the CA" >&2
    exit 1
fi

umask 077
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:4096 -out "$OUT/myCA.key"
umask 022

openssl req -x509 -new -key "$OUT/myCA.key" -sha256 -days "$DAYS" \
    -subj "/CN=${CN}/O=safe-import" \
    -addext "basicConstraints=critical,CA:TRUE,pathlen:0" \
    -addext "keyUsage=critical,keyCertSign,cRLSign" \
    -addext "subjectKeyIdentifier=hash" \
    -out "$OUT/myCA.pem"

chmod 600 "$OUT/myCA.key"
chmod 644 "$OUT/myCA.pem"

echo "CA created:"
echo "  certificate: $OUT/myCA.pem  (distribute to inner clients)"
echo "  private key: $OUT/myCA.key  (keep secret, readable only by the si-inner user)"
echo
echo "Configure si-inner with:"
echo "  SI_INNER_CA_CERT=$OUT/myCA.pem"
echo "  SI_INNER_CA_KEY=$OUT/myCA.key"
openssl x509 -in "$OUT/myCA.pem" -noout -subject -enddate -fingerprint -sha256
