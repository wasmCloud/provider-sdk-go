package provider

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/nats-io/nkeys"

	core "github.com/wasmcloud/interfaces/core/tinygo"
)

func EncodeClaims(i *core.Invocation, hostData core.HostData, guid string) error {
	var Ed25519SigningMethod jwt.SigningMethodEd25519
	jwt.RegisterSigningMethod("Ed25519", func() jwt.SigningMethod { return &Ed25519SigningMethod })

	service, err := nkeys.FromSeed([]byte(hostData.InvocationSeed))
	if err != nil {
		return err
	}

	pkey, err := service.PrivateKey()
	if err != nil {
		return err
	}

	pubkey, err := service.PublicKey()
	if err != nil {
		return err
	}

	rKey, err := nkeys.Decode(nkeys.PrefixBytePrivate, pkey)
	if err != nil {
		return err
	}

	contract := strings.ReplaceAll(string(i.Origin.ContractId), ":", "/")
	claims := Claims{
		StandardClaims: jwt.StandardClaims{
			IssuedAt: time.Now().Unix(),
			Issuer:   pubkey,
			Subject:  guid,
		},
		ID: guid,
		Wascap: Wascap{
			TargetURL: "wasmbus://" + i.Target.PublicKey + "/" + i.Operation,
			OriginURL: "wasmbus://" + contract + "/" + hostData.LinkName + "/" + i.Origin.PublicKey,
		},
	}

	var priKey ed25519.PrivateKey = rKey

	var b bytes.Buffer
	b.WriteString(claims.Wascap.OriginURL)
	b.WriteString(claims.Wascap.TargetURL)
	b.WriteString(i.Operation)
	b.WriteString(string(i.Msg))
	hash := sha256.Sum256(b.Bytes())
	claims.Wascap.Hash = strings.ToUpper(hex.EncodeToString(hash[:]))

	token := jwt.NewWithClaims(&Ed25519SigningMethod, claims)
	token.Header["alg"] = "Ed25519"
	token.Header["typ"] = "jwt"

	jwtstring, err := token.SignedString(priKey)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}
	i.EncodedClaims = jwtstring

	return nil
}
