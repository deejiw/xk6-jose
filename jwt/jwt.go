// MIT License
//
// Copyright (c) 2021 Iván Szkiba
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
)

type Module struct{}

func New() *Module {
	return &Module{}
}

var ErrUnsupportedKey = errors.New("unsupported key")

func (m *Module) Sign(key *jose.JSONWebKey, payload, header map[string]interface{}) (string, error) {
	opts := &jose.SignerOptions{}
	opts = opts.WithType("JWT")

	for k, v := range header {
		opts.WithHeader(jose.HeaderKey(k), v)
	}

	sig, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.SignatureAlgorithm(key.Algorithm), Key: key}, opts)
	if err != nil {
		return "", err
	}

	str, err := jwt.Signed(sig).Claims(payload).CompactSerialize()
	if err != nil {
		return "", err
	}

	return str, nil
}

func (m *Module) SignRSA(key *rsa.PrivateKey, claim interface{}) (string, error) {
	var signerOpts = jose.SignerOptions{}
	signerOpts.WithType("JWT")

	joseSigner, err := jose.NewSigner(jose.SigningKey{
		Algorithm: jose.RS256,
		Key:       key,
	}, &signerOpts)
	if err != nil {
		return "", fmt.Errorf("failed to create signer:%+v", err)
	}
	token, err := jwt.Signed(joseSigner).Claims(claim).CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("failed to sign:%+v", err)
	}
	return token, nil
}

func (m *Module) Decode(compact string) (interface{}, error) {
	token, err := jwt.ParseSigned(compact)
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{}

	if err := token.UnsafeClaimsWithoutVerification(&payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (m *Module) Verify(compact string, keys ...interface{}) (interface{}, error) {
	token, err := jwt.ParseSigned(compact)
	if err != nil {
		return nil, err
	}

	set := make([]jose.JSONWebKey, len(keys))

	for _, k := range keys {
		switch k.(type) {
		case jose.JSONWebKey:
			set = append(set, k.(jose.JSONWebKey))
		case *jose.JSONWebKey:
			set = append(set, *k.(*jose.JSONWebKey))
		case *jose.JSONWebKeySet:
			set = append(set, k.(*jose.JSONWebKeySet).Keys...)
		default:
			return nil, fmt.Errorf("%w: %T %v", ErrUnsupportedKey, k, k)
		}
	}

	payload := map[string]interface{}{}

	if err := token.Claims(&jose.JSONWebKeySet{Keys: set}, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}
