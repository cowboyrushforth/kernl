package models
import "fmt"
import "github.com/cowboyrushforth/gosalmon"
import "strings"
import "crypto/cipher"
import "crypto/aes"
import "crypto/x509"
import "crypto/rsa"
//import "crypto/sha1"
import "crypto/rand"
import "encoding/base64"
import "encoding/pem"
import "net/http"
import "net/url"

// sends a 'start sharing' notification from
// the owner of this person entry to the person 
// described in it
func SendSharingNotification(user *User, person *Person)  (resp *http.Response, err error) {
  template := `<XML>
  <post>
    <request>
      <sender_handle>$sender</sender_handle>
      <recipient_handle>$receipient</sender_handle>
    </request>
  </post>
</XML>`

  payload := strings.Replace(template, "$sender", user.AccountIdentifier, 1)
  payload = strings.Replace(payload, "$recipient", person.AccountIdentifier, 1)
  encryption_header, inner_iv, inner_key := generateEncryptionHeader(user, person)
  encrypted_payload := generateEncryptedPayload(payload, inner_iv, inner_key)

  salmon := gosalmon.Salmon{
    EncryptionHeader: encryption_header,
    Payload: base64.URLEncoding.EncodeToString([]byte(encrypted_payload)),
    Datatype: "application/atom+xml",
    Algorithm: "RSA-SHA256",
    Encoding: "base64url",
    RSAKey: user.RSAKey,
  }
  xml := salmon.EncodeToXml()
  salmon_endpoint := person.PodUrl + "/receive/users/" + person.RemoteGuid
  return sendPreparedSalmon(xml, salmon_endpoint)
}

func sendPreparedSalmon(xml string, salmon_endpoint string) (resp *http.Response, err error) {
  v := url.Values{}
  v.Set("xml", xml)
  v.Encode()
  return http.PostForm(salmon_endpoint, v)
}

func generateEncryptedPayload(payload string, inner_iv []byte, inner_key []byte) string {
  // sanity check
  for (len(payload)%aes.BlockSize != 0) { 
    payload = payload + "0"
  }
  block, block_err := aes.NewCipher(inner_key)
  if block_err != nil {
    panic(block_err)
  }
  enc_payload := make([]byte, aes.BlockSize+len(payload))
  header_mode := cipher.NewCBCEncrypter(block, inner_iv)
  header_mode.CryptBlocks(enc_payload, []byte(payload))
  return base64.StdEncoding.EncodeToString(enc_payload)
}

func generateEncryptionHeader(user *User, person *Person) (string, []byte, []byte) {
  

  template := `<decrypted_header>
  <iv>$inner_iv</iv>
  <aes_key>$inner_key</aes_key>
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


  // sanity check
  for (len(dec_header)%aes.BlockSize != 0) { 
    dec_header = dec_header + "0"
  }

  // make block
  header_block, header_err := aes.NewCipher(outer_key)
  if header_err != nil {
    panic(header_err)
  }
  // encrypt header
  enc_header := make([]byte, aes.BlockSize+len(dec_header))
  header_mode := cipher.NewCBCEncrypter(header_block, outer_iv)
  header_mode.CryptBlocks(enc_header, []byte(dec_header))
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
  braw, _ := base64.StdEncoding.DecodeString(person.RSAPubKey)
  fmt.Println("wheee")
  fmt.Println(person.RSAPubKey)
  fmt.Println(string(braw))
  fmt.Println("wheee")
  p, _ := pem.Decode(braw)
  if p == nil {
    panic("could not parse public key")
  }
  pubkey, pubkeyerr := x509.ParsePKIXPublicKey(p.Bytes)
  if(pubkeyerr != nil) {
    panic("could not parse public key")
  }
  //hash := sha1.New()

  // NEED PKCS1_PADDING HERE
   enc_outer_bundle, outer_bundle_err := rsa.EncryptPKCS1v15(rand.Reader, pubkey.(*rsa.PublicKey), []byte(outer_bundle))
//  enc_outer_bundle, outer_bundle_err := rsa.EncryptOAEP(hash, rand.Reader, pubkey.(*rsa.PublicKey), []byte(outer_bundle), nil)
  if outer_bundle_err != nil {
    panic(outer_bundle_err)
  }

  encrypted_header_json_object := `{
  "aes_key": "$aes_key",
  "ciphertext": "$ciphertext" 
}`

  encrypted_header_json_object = strings.Replace(encrypted_header_json_object, "$aes_key",
                                                 base64.StdEncoding.EncodeToString(enc_outer_bundle), 1)
  encrypted_header_json_object = strings.Replace(encrypted_header_json_object, "$ciphertext",
                                                 base64.StdEncoding.EncodeToString(enc_header), 1)


  encrypted_header := "<encrypted_header>"+base64.StdEncoding.EncodeToString([]byte(encrypted_header_json_object))+"</encrypted_header>"

  return encrypted_header, inner_iv, inner_key
}
