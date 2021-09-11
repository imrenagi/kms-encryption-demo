package payment

import (
  b64 "encoding/base64"

  "github.com/google/tink/go/aead"
  "github.com/google/tink/go/core/registry"
  "github.com/google/tink/go/integration/gcpkms"
  "github.com/google/tink/go/keyset"
  "github.com/google/tink/go/tink"
  "github.com/rs/zerolog/log"
)

var (
  encrypter tink.AEAD
)

const (
  keyURI = "gcp-kms://projects/imre-cloud-kms-test/locations/global/keyRings/payments/cryptoKeys/db-credit-card"
)

func init() {
  var err error
  gcpclient, err := gcpkms.NewClient(keyURI)
  if err != nil {
    log.Fatal().Msg(err.Error())
  }
  registry.RegisterKMSClient(gcpclient)

  dek := aead.AES128CTRHMACSHA256KeyTemplate()
  kh, err := keyset.NewHandle(aead.KMSEnvelopeAEADKeyTemplate(keyURI, dek))
  if err != nil {
    log.Fatal().Msg(err.Error())
  }

  encrypter, err = aead.New(kh)
  if err != nil {
    log.Fatal().Msg(err.Error())
  }
}

func encrypt(value []byte, associatedData []byte) string {
  b, err := encrypter.Encrypt(value, associatedData)
  if err != nil {
    log.Fatal().Msg(err.Error())
  }

  return b64.StdEncoding.EncodeToString(b)
}

func decrypt(decodedValue string, associatedData []byte) string {
  decoded, err := b64.StdEncoding.DecodeString(decodedValue)
  if err != nil {
    log.Fatal().Msg(err.Error())
  }

  b, err := encrypter.Decrypt(decoded, associatedData)
  if err != nil {
    log.Fatal().Msg(err.Error())
  }
  return string(b)
}