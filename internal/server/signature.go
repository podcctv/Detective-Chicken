package server

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var signatureInputPattern = regexp.MustCompile(`created=(\d+);keyid="([^"]+)";nonce="([^"]+)"`)

type signatureMeta struct {
	Created      int64
	KeyID, Nonce string
}

func parseSignatureInput(value string) (signatureMeta, error) {
	m := signatureInputPattern.FindStringSubmatch(value)
	if len(m) != 4 {
		return signatureMeta{}, errors.New("invalid Signature-Input")
	}
	created, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return signatureMeta{}, err
	}
	return signatureMeta{Created: created, KeyID: m[2], Nonce: m[3]}, nil
}

func contentDigest(body []byte) string {
	sum := sha256.Sum256(body)
	return "sha-256=:" + base64.StdEncoding.EncodeToString(sum[:]) + ":"
}

func signingBase(method, path, digest string, meta signatureMeta) []byte {
	return []byte(fmt.Sprintf("%s\n%s\n%s\n%d\n%s\n%s", strings.ToUpper(method), path, digest, meta.Created, meta.Nonce, meta.KeyID))
}

func verifySignature(r *http.Request, body, publicKey []byte, meta signatureMeta) error {
	if len(publicKey) != ed25519.PublicKeySize {
		return errors.New("invalid public key")
	}
	if delta := time.Now().Unix() - meta.Created; delta > 300 || delta < -300 {
		return errors.New("signature timestamp outside allowed window")
	}
	expected := contentDigest(body)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(r.Header.Get("Content-Digest"))) != 1 {
		return errors.New("content digest mismatch")
	}
	value := r.Header.Get("Signature")
	value = strings.TrimPrefix(value, "sig1=:")
	value = strings.TrimSuffix(value, ":")
	sig, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return errors.New("invalid signature encoding")
	}
	if !ed25519.Verify(publicKey, signingBase(r.Method, r.URL.Path, expected, meta), sig) {
		return errors.New("invalid signature")
	}
	return nil
}
