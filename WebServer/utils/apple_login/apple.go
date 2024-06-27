package apple_login

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"math/big"
	"net/http"
	"strings"
)

func NewAppleLogin(clientId string) Apple {
	return Apple{
		clientId: clientId,
		jwksUrl:  "https://appleid.apple.com/auth/keys",
	}
}

type Apple struct {
	clientId string
	jwksUrl  string
}

func (p *Apple) VerifyIdToken(idToken string) (jwt.MapClaims, error) {
	if idToken == "" {
		return nil, errors.New("empty id_token")
	}

	// extract the token header params and claims
	claims := jwt.MapClaims{}
	t, _, err := jwt.NewParser().ParseUnverified(idToken, claims)
	if err != nil {
		return nil, err
	}

	// validate common claims per https://developer.apple.com/documentation/sign_in_with_apple/sign_in_with_apple_rest_api/verifying_a_user#3383769
	isu, err := claims.GetIssuer()
	if err != nil {
		return nil, err
	}
	if isu != "https://appleid.apple.com" {
	}
	adu, err := claims.GetAudience()
	if err != nil {
		return nil, err
	}
	if len(adu) <= 0 {
		return nil, errors.New("adu is empty")
	}
	var isInAdu bool
	for _, v := range adu {
		if v == p.clientId {
			isInAdu = true
		}
	}
	if !isInAdu {
		return nil, errors.New("check adu err")
	}

	// fetch the public key set
	kid, _ := t.Header["kid"].(string)
	if kid == "" {
		return nil, errors.New("missing kid header value")
	}

	key, err := p.fetchJWK(kid)
	if err != nil {
		return nil, err
	}

	// decode the key params per RFC 7518 (https://tools.ietf.org/html/rfc7518#section-6.3)
	// and construct a valid publicKey from them
	exponent, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(key.E, "="))
	if err != nil {
		return nil, err
	}

	modulus, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(key.N, "="))
	if err != nil {
		return nil, err
	}

	publicKey := &rsa.PublicKey{
		// https://tools.ietf.org/html/rfc7517#appendix-A.1
		E: int(big.NewInt(0).SetBytes(exponent).Uint64()),
		N: big.NewInt(0).SetBytes(modulus),
	}

	// verify the id_token
	parser := jwt.NewParser(jwt.WithValidMethods([]string{key.Alg}))

	parsedToken, err := parser.Parse(idToken, func(t *jwt.Token) (any, error) {
		return publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		return claims, nil
	}

	return nil, errors.New("the parsed id_token is invalid")
}

type jwk struct {
	Kty string
	Kid string
	Use string
	Alg string
	N   string
	E   string
}

func (p *Apple) fetchJWK(kid string) (*jwk, error) {
	req, err := http.NewRequestWithContext(context.TODO(), "GET", p.jwksUrl, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// http.Client.Get doesn't treat non 2xx responses as error
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf(
			"failed to verify the provided id_token (%d):\n%s",
			res.StatusCode,
			string(rawBody),
		)
	}

	jwks := struct {
		Keys []*jwk
	}{}
	if err := json.Unmarshal(rawBody, &jwks); err != nil {
		return nil, err
	}

	for _, key := range jwks.Keys {
		if key.Kid == kid {
			return key, nil
		}
	}

	return nil, fmt.Errorf("jwk with kid %q was not found", kid)
}
