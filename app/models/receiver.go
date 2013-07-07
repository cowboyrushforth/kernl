package models

// receiver parses all bits
// needed from salmon transmission
// decrypts, and verifies payload
// is sent by who it says it was

import "github.com/robfig/revel"
import "github.com/cowboyrushforth/gosalmon"
import "github.com/garyburd/redigo/redis"
import "encoding/base64"
import "encoding/json"
import "encoding/xml"
import "crypto/rsa"
import "encoding/pem"
import "crypto/x509"
import "crypto/rand"
import "crypto/aes"
import "crypto/cipher"
import "errors"

type OuterPackage struct {
  AesKey string `json:"aes_key"`
  Ciphertext string `json:"ciphertext"`
}

type InnerPackage struct {
  Key string `json:"key"`
  Iv string `json:"iv"`
}

type DecryptedHeader struct {
  AesKey   string `xml:"aes_key"`
  Iv       string `xml:"iv"`
  AuthorId string `xml:"author_id"`
}

func ParseVerifiedSalmonPayload(rc redis.Conn, user *User, xmlstr string) (sender *Person, payload string, err error) {
  salmon := gosalmon.Salmon{}
  errs := salmon.DecodeFromXml(xmlstr)
  if errs != nil {
    panic("Salmon Decoding Problem")
  }
  encrypted_header := salmon.EncryptionHeader
  raw, _ := base64.URLEncoding.DecodeString(encrypted_header)

  // decrypt aes_key with
  var outer_pkg OuterPackage
  errj := json.Unmarshal(raw, &outer_pkg)
  if errj != nil {
    panic(errj)
  }

   p, _ := pem.Decode([]byte(user.RSAKey))
   if p == nil {
     panic("could not parse private key")
   }
   pk, err := x509.ParsePKCS1PrivateKey(p.Bytes)

   raw_aes_key, _ := base64.StdEncoding.DecodeString(outer_pkg.AesKey)
   result, errd := rsa.DecryptPKCS1v15(rand.Reader, pk, raw_aes_key) 
   if errd != nil {
        panic(errd)
   }

   var inner_pkg InnerPackage
   _ = json.Unmarshal(result, &inner_pkg)

   // now finally decrypt the header with the inner_pkg
   raw_key, _ := base64.StdEncoding.DecodeString(inner_pkg.Key)
   raw_iv, _  := base64.StdEncoding.DecodeString(inner_pkg.Iv)
   header_payload, _ := base64.StdEncoding.DecodeString(outer_pkg.Ciphertext)
   block, err := aes.NewCipher(raw_key)
   if err != nil {
     panic(err)
   }
   if len(payload)%aes.BlockSize != 0 {
     panic("payload is not a multiple of the block size")
   }
   mode := cipher.NewCBCDecrypter(block, raw_iv)
   mode.CryptBlocks(header_payload, header_payload)

   var header DecryptedHeader
   xml.Unmarshal(header_payload, &header)

   revel.INFO.Println("to", user.AccountIdentifier, "from", header.AuthorId)

   // now that we have seen who its from
   // get our versio of that public key
   // and verify the salmon sig
   person, person_err := PersonFromUid(rc, "person:"+header.AuthorId)
   if person_err != nil {
     // we appear to not have this person.
     // try to finger them.
     person_err = nil
     person, person_err = PersonFromWebFinger(header.AuthorId)
     if person_err != nil {
       panic("can not locate person")
     }
     person.Insert(rc)
   }
   

   salmon.RSAPubKey = person.RSAPubKey
   if salmon.IsValid() {
     // ok decrypt the final payload
     ppayload, _ := base64.StdEncoding.DecodeString(salmon.Payload)
     pkey, _ := base64.StdEncoding.DecodeString(header.AesKey)
     piv, _  := base64.StdEncoding.DecodeString(header.Iv)
     pblock, perr := aes.NewCipher(pkey)
     if perr != nil {
       panic(perr)
     }
     if len(ppayload)%aes.BlockSize != 0 {
         panic("payload is not a multiple of the block size")
     }
     mode := cipher.NewCBCDecrypter(pblock, piv)
     mode.CryptBlocks(ppayload, ppayload)
     return person, string(ppayload), nil
   }
   return nil, "", errors.New("Salmon Not Verified")
}

func ParsePublicVerifiedSalmonPayload(rc redis.Conn, xmlstr string) (sender *Person, payload string, err error) {
  salmon := gosalmon.Salmon{}
  errs := salmon.DecodeFromXml(xmlstr)
  if errs != nil {
    panic("Salmon Decoding Problem")
  }
  revel.INFO.Println("from", salmon.AuthorId)

   // now that we have seen who its from
   // get our versio of that public key
   // and verify the salmon sig
   person, person_err := PersonFromUid(rc, "person:"+salmon.AuthorId)
   if person_err != nil {
     // we appear to not have this person.
     // try to finger them.
     person_err = nil
     person, person_err = PersonFromWebFinger(salmon.AuthorId)
     if person_err != nil {
       panic("can not locate person")
     }
     person.Insert(rc)
   }
   
   salmon.RSAPubKey = person.RSAPubKey
   if salmon.IsValid() {
     return person, salmon.Payload, nil
   }
   return nil, "", errors.New("Salmon Not Verified")
}
