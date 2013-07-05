package controllers

import "github.com/robfig/revel"
import "github.com/cowboyrushforth/gosalmon"
import "kernl/app/models"
import "net/url"
import "encoding/base64"
import "encoding/json"
import "encoding/xml"
import "crypto/rsa"
import "encoding/pem"
import "crypto/x509"
import "crypto/rand"
import "crypto/aes"
import "crypto/cipher"

type Salmon struct {
  Kernl
}

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

func (c Salmon) Receive() revel.Result {

  guid := c.Params.Get("guid")
  raw_xml := c.Params.Get("xml")

  // get redis handle
  rc := GetRedisConn()
  defer rc.Close()

  user, err := models.UserFromGuid(rc, guid)
  if err != nil {
    return c.NotFound("User Not Found")
  }

  revel.INFO.Println("received salmon slap for guid", guid)
  revel.INFO.Println("Raw XML")
  sane_xml, _ := url.QueryUnescape(raw_xml)
  salmon := gosalmon.Salmon{}
  errs := salmon.DecodeFromXml(sane_xml)
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
   payload, _ := base64.StdEncoding.DecodeString(outer_pkg.Ciphertext)
   block, err := aes.NewCipher(raw_key)
   if err != nil {
     panic(err)
   }
   if len(payload)%aes.BlockSize != 0 {
     panic("payload is not a multiple of the block size")
   }
   mode := cipher.NewCBCDecrypter(block, raw_iv)
   mode.CryptBlocks(payload, payload)

   var header DecryptedHeader
   xml.Unmarshal(payload, &header)

   revel.INFO.Println("to", user.AccountIdentifier, "from", header.AuthorId)

   // now that we have seen who its from
   // get our versio of that public key
   // and verify the salmon sig
   person, _ := models.PersonFromUid(rc, header.AuthorId)
   salmon.RSAPubKey = person.RSAPubKey
   if salmon.IsValid() {
     revel.INFO.Println("SALMON VALID!")
   } else {
     revel.INFO.Println("SALMON NOT VALID!")
   }
   
   return c.NotFound("oops")
}
