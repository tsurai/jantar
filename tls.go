package jantar

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"time"
)

// Additional TLS cipher
const (
	TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384 uint16 = 0xc02c
	TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384   uint16 = 0xc030
)

var tlsConfig = &tls.Config{
	CipherSuites: []uint16{
		TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA256,
		tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
		tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
		tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
	},
	PreferServerCipherSuites: true,
	MinVersion:               tls.VersionTLS10,
}

func loadTLSCertificate(config *TLSConfig) error {
	var err error
	var cert tls.Certificate
	var certPem = config.CertPem
	var keyPem = config.KeyPem

	if config.CertFile != "" {
		certPem, err = ioutil.ReadFile(config.CertFile)
		if err != nil {
			return err
		}
	}

	if config.KeyFile != "" {
		if keyPem, err = ioutil.ReadFile(config.KeyFile); err != nil {
			return err
		}
	}

	if certPem == nil {
		return errors.New("no certificate pem block given")
	}

	if keyPem == nil {
		return errors.New("no key pem block given")
	}

	if cert, err = tls.X509KeyPair(certPem, keyPem); err != nil {
		return err
	}

	if cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0]); err != nil {
		return err
	}

	checkTLSCertificate(cert.Leaf)

	config.cert = cert

	return nil
}

func checkTLSCertificate(cert *x509.Certificate) {
	// is pre heartbleed
	if cert.NotBefore.Before(time.Date(2014, time.April, 07, 12, 0, 0, 0, time.UTC)) {
		Log.Warningd(JLData{"issued": cert.NotBefore.UTC().Format(time.RFC822), "fixed": "07 April 14"}, "x509 certificate has been issued before heartbleed(CVE-2014-0160) fix!\nYour secret key and other private information might have been leaked")
	}
}
