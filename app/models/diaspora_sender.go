package models

// sender prepared all bits
// needed for salmon transmission
// and sends to diaspora

import "github.com/cowboyrushforth/gosalmon"
import "strings"
import "crypto/cipher"
import "crypto/aes"
import "crypto/x509"
import "crypto/rsa"
import "crypto/rand"
import "encoding/base64"
import "encoding/pem"
import "net/http"
import "net/url"
import "time"

// sends a 'start sharing' notification from
// the owner of this person entry to the person 
// described in it
func Diaspora_SendSharingNotification(user *User, person *Person)  (resp *http.Response, err error) {
  template := `<XML>
  <post>
    <request>
      <sender_handle>$sender</sender_handle>
      <recipient_handle>$recipient</recipient_handle>
    </request>
  </post>
</XML>`

  payload := strings.Replace(template, "$sender", user.AccountIdentifier, 1)
  payload = strings.Replace(payload, "$recipient", person.AccountIdentifier, 1)
  return sendSalmon(payload, user, person)
}

func Diaspora_SendStatusMessage(user *User, person *Person, post *Post) (resp *http.Response, err error) {
        template := `<XML>
  <post>
    <status_message>
      <raw_message>$status_message</raw_message>
      <guid>$guid</guid>
      <diaspora_handle>$sender</diaspora_handle>
      <public>$public</public>
      <created_at>$timestamp</created_at>
    </status_message>
  </post>
</XML>`

  payload := strings.Replace(template, "$status_message", post.Message, 1)
  payload = strings.Replace(payload, "$guid", post.Guid, 1)
  payload = strings.Replace(payload, "$sender", user.AccountIdentifier, 1)
  payload = strings.Replace(payload, "$timestamp", time.Unix(post.CreatedAt, 0).Format("2006-01-02 15:04:05 MST"), 1)

  post.Public = false // FIXME
  if post.Public == true {
    payload = strings.Replace(payload, "$public", "true", 1)
  } else {
    payload = strings.Replace(payload, "$public", "false", 1)
  }

  payload = strings.Replace(payload, "$sender", user.AccountIdentifier, 1)
  payload = strings.Replace(payload, "$recipient", person.AccountIdentifier, 1)

  return sendSalmon(payload, user, person)
}

func sendSalmon(payload string, user *User, person *Person) (resp *http.Response, err error) {
  encryption_header, inner_iv, inner_key := generateEncryptionHeader(user, person)
  encrypted_payload := generateEncryptedPayload(payload, inner_iv, inner_key)

  salmon := gosalmon.Salmon{
    EncryptionHeader: encryption_header,
    Payload: encrypted_payload,
    Datatype: "application/atom+xml",
    Algorithm: "RSA-SHA256",
    Encoding: "base64url",
    RSAKey: user.RSAKey,
  }
  xml := salmon.EncodeToXml(true)
  salmon_endpoint := person.PodUrl + "/receive/users/" + person.RemoteGuid
  return sendPreparedSalmon(xml, salmon_endpoint)
}

func sendPreparedSalmon(xml string, salmon_endpoint string) (resp *http.Response, err error) {
  v := url.Values{}
  v.Set("xml", url.QueryEscape(xml))
  v.Encode()
  return http.PostForm(salmon_endpoint, v)
}

func generateEncryptedPayload(payload string, iv []byte, key []byte) string {
  ciph, err := aes.NewCipher(key)
  if err != nil {
    panic(err)
  }
  encryptor := cipher.NewCBCEncrypter(ciph, iv)
  pad := encryptor.BlockSize() - len(payload)%encryptor.BlockSize()
  enc_payload := make([]byte, len(payload), len(payload)+pad)
  copy(enc_payload, payload)
  for i := 0; i < pad; i++ {
    enc_payload = append(enc_payload, byte(pad))
  }
  encryptor.CryptBlocks(enc_payload, enc_payload)
  return base64.StdEncoding.EncodeToString(enc_payload)
}

func generateEncryptionHeader(user *User, person *Person) (string, []byte, []byte) {
  
  template := `<decrypted_header>
  <iv>$inner_iv</iv>
  <aes_key>$inner_key</aes_key>
  <author_id>$author_id</author_id>
  <author>
    <name>$display_name</name>
    <uri>$identifier</uri>
  </author>
</decrypted_header>`

  inner_key := []byte(RandomSHA256()[0:32])
  inner_iv  := []byte(RandomString(16))
  outer_key := []byte(RandomSHA256()[0:32])
  outer_iv  := []byte(RandomString(16))

  dec_header := strings.Replace(template, "$inner_iv", base64.StdEncoding.EncodeToString(inner_iv), 1)
  dec_header = strings.Replace(dec_header, "$inner_key", base64.StdEncoding.EncodeToString(inner_key), 1)
  dec_header = strings.Replace(dec_header, "$display_name", user.DisplayName, 1)
  dec_header = strings.Replace(dec_header, "$identifier", user.AccountIdentifier, 1)
  dec_header = strings.Replace(dec_header, "$author_id", user.AccountIdentifier, 1)

  // encrypt header
  enc_header := generateEncryptedPayload(dec_header, outer_iv, outer_key)

  // make outer aes key bundle
  outer_bundle := `{
  "iv": "$outer_iv",
  "key": "$outer_key"
}`

  // fill in template 
  outer_bundle = strings.Replace(outer_bundle, "$outer_iv", 
                                 base64.StdEncoding.EncodeToString(outer_iv), 1)
  outer_bundle = strings.Replace(outer_bundle, "$outer_key", 
                                 base64.StdEncoding.EncodeToString(outer_key), 1)

  // encrypt outer bundle using recipients public key
  p, _ := pem.Decode([]byte(person.RSAPubKey))
  if p == nil {
    panic("could not parse public key")
  }
  pubkey, pubkeyerr := x509.ParsePKIXPublicKey(p.Bytes)
  if(pubkeyerr != nil) {
    panic("could not parse public key")
  }

  enc_outer_bundle, outer_bundle_err := rsa.EncryptPKCS1v15(rand.Reader, pubkey.(*rsa.PublicKey), []byte(outer_bundle))
  if outer_bundle_err != nil {
    panic(outer_bundle_err)
  }

  enc_json := `{"aes_key": "$aes_key", "ciphertext": "$ciphertext" }`
  enc_json = strings.Replace(enc_json, "$aes_key",
                             base64.StdEncoding.EncodeToString(enc_outer_bundle), 1)
  enc_json = strings.Replace(enc_json, "$ciphertext", enc_header, 1)
  encrypted_header := "<encrypted_header>"+base64.StdEncoding.EncodeToString([]byte(enc_json))+"</encrypted_header>"
  return encrypted_header, inner_iv, inner_key
}
