package amber

import (
  "fmt"
  "time"
  "net/http"
  "net/url"
  "crypto/aes"
  "crypto/sha256"
  "crypto/hmac"
  "crypto/cipher"
  "crypto/rand"
  "encoding/base64"
)

var (
  secretkey []byte
)

func SecureCookie(usertoken string, data string) *http.Cookie {
  // set the expiration date to one year in the future
  expiration := time.Now().AddDate(1, 0, 0).Unix()

  // compute encryption key with HMAC(usertoken|expiration, secretkey)
  mac := hmac.New(sha256.New, secretkey)
  mac.Write([]byte(usertoken + string(expiration)))
  key := mac.Sum(nil)

  // encrypt cookie data block
  encdata := encrypt(data, key)

  // compute hmac hash with HMAC(usertoken|expiration|data, key)
  mac = hmac.New(sha256.New, key)
  mac.Write([]byte(usertoken + string(expiration) + data))
  hhmac := mac.Sum(nil)

  // combine all data and return a secure cookie
  tmp := fmt.Sprintf("%s %d %x %x", usertoken, expiration, encdata, hhmac)
  return &http.Cookie{Name: "AMBER_SESSION", Value: url.QueryEscape(tmp), Secure: false, HttpOnly: true, Path: "/"}
}

func UnlockCookie(cookie *http.Cookie) *http.Cookie {
  var usertoken string
  var expiration int64
  var encdata, cookieMAC []byte

  // parse values from cookie
  value, _ := url.QueryUnescape(cookie.Value)
  if n, err := fmt.Sscanf(value, "%s %d %x %x", &usertoken, &expiration, &encdata, &cookieMAC); n != 4 || err != nil {
    return nil
  }

  // check if cookie is still valid
  if expiration < time.Now().Unix() {
    return nil
  }

  // compute decryption key with HMAC(usertoken|expiration, secretkey)
  mac := hmac.New(sha256.New, secretkey)
  mac.Write([]byte(usertoken + string(expiration)))
  key := mac.Sum(nil)

  // decrypt cookie data block
  data := decrypt(encdata, key)

  // compute expected hmac with HMAC(usertoken|expiration|data, key)
  mac = hmac.New(sha256.New, key)
  mac.Write([]byte(usertoken + string(expiration) + data))

  // compare both hmac hashes
  if hmac.Equal(cookieMAC, mac.Sum(nil)) {
    cookie.Value = usertoken + " " + data
    return cookie
  }

  return nil
}

func encrypt(plaintext string, key []byte) []byte {
  // create a new aes block cipher with a given key
  block, err := aes.NewCipher(key)
  if err != nil {
    return nil
  }

  // base64 encode the plaintext and generate a random IV
  baseplaintext := base64.StdEncoding.EncodeToString([]byte(plaintext))
  ciphertext := make([]byte, aes.BlockSize + len(baseplaintext))
  iv := ciphertext[:aes.BlockSize]
  
  if _, err := rand.Read(iv); err != nil {
    return nil
  }
  
  // aes encrypt the base64 encoded plaintext
  cfb := cipher.NewCFBEncrypter(block, iv)
  cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(baseplaintext))
  
  return ciphertext
}

func decrypt(data []byte, key []byte) string {
  // create a new aes block cipher with a given key
  block, err := aes.NewCipher(key)
  if err != nil {
    return ""
  }

  // check if the size is valid
  if len(data) < aes.BlockSize {
    return ""
  }
  
  // fetch the IV and data
  iv := data[:aes.BlockSize]
  data = data[aes.BlockSize:]

  // decrypt the data
  cfb := cipher.NewCFBDecrypter(block, iv)
  cfb.XORKeyStream(data, data)

  // decode the base64 encoded data
  data, err = base64.StdEncoding.DecodeString(string(data))
  if err != nil {
    return ""
  }

   return string(data)
}

// WARNING: this will render all cookies invalid after an application restart
func init() {
  // generate a 256 byte secretkey on startup
  secretkey := make([]byte, 32)
  if _, err := rand.Read(secretkey); err != nil {
    logger.Fatal("[Fatal] Failed to generate secret key.", err)
  }
}