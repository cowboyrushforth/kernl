package models
import "math/big"
import "encoding/hex"
import "crypto/rand"
import "crypto/sha256"
import "strings"
import "github.com/robfig/revel"

func RandomString(n int) string {
  const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
  symbols := big.NewInt(int64(len(alphanum)))
  states := big.NewInt(0)
  states.Exp(symbols, big.NewInt(int64(n)), nil)
  r, err := rand.Int(rand.Reader, states)
  if err != nil {
    panic(err)
  }
  var bytes = make([]byte, n)
  r2 := big.NewInt(0)
  symbol := big.NewInt(0)
  for i := range bytes {
    r2.DivMod(r, symbols, symbol)
    r, r2 = r2, r
    bytes[i] = alphanum[symbol.Int64()]
  }
  return string(bytes)
}

func RandomSHA256() string {
  hash := sha256.New()
  hash.Write([]byte(RandomString(64)))
  md := hash.Sum(nil)
  mdStr := hex.EncodeToString(md)
  return mdStr
}

// takes a given word (string)
// and computers a score used
// to sort on
func WordScore(word string) int {
  pos := 3
  score := 0
  for i:= range []byte(strings.ToUpper(word)) {
    if pos < 0 {
      break
    }
    ns := i*(256*pos)
    score += ns
    pos -= 1
  }
  return score
}

func IsLocalIdentifier(identifier string) bool {
  host_suffix := revel.Config.StringDefault("host.suffix", "localhost:9000")
  l := len(host_suffix)
  q := len(identifier) - (l+1)
  return identifier[q:] == "@"+host_suffix
}
