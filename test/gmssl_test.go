/*
 * Copyright 2020 The Hyperledger-TWGC Project Authors. All Rights Reserved.
 *
 * Licensed under the Apache License 2.0 (the "License").  You may not use
 * this file except in compliance with the License.  You can obtain a copy
 * in the file LICENSE in the source distribution or at
 * https://www.openssl.org/source/license.html
 */
package main

import (
	"fmt"
	"github.com/Hyperledger-TWGC/Gm-Go/gmssl"
	"github.com/Hyperledger-TWGC/Gm-Go/gmssl/sm3"
)

func main() {

	engines := gmssl.GetEngineNames()
	/* SM3 digest with GmSSL-Go API */
	sm3ctx, _ := gmssl.NewDigestContext("SM3")
	sm3ctx.Update([]byte("a"))
	sm3ctx.Update([]byte("bc"))
	sm3digest, _ := sm3ctx.Final()
	fmt.Printf("sm3(\"abc\") = %x\n", sm3digest)

	/* SM3 digest with Go hash.Hash API */
	sm3hash := sm3.New()
	sm3hash.Write([]byte("abc"))
	fmt.Printf("sm3(\"abc\") = %x\n", sm3hash.Sum(nil))

	/* HMAC-SM3 */
	hmac_sm3, _ := gmssl.NewHMACContext("SM3", []byte("this is the key"))
	hmac_sm3.Update([]byte("ab"))
	hmac_sm3.Update([]byte("c"))
	mactag, _ := hmac_sm3.Final()
	fmt.Printf("hmac-sm3(\"abc\") = %x\n", mactag)

	/* Generate random key and IV */
	keylen, _ := gmssl.GetCipherKeyLength("SMS4")
	key, _ := gmssl.GenerateRandom(keylen)
	ivlen, _ := gmssl.GetCipherIVLength("SMS4")
	iv, _ := gmssl.GenerateRandom(ivlen)

	/* SMS4-CBC Encrypt/Decrypt */
	encryptor, _ := gmssl.NewCipherContext("SMS4", key, iv, true)
	ciphertext1, _ := encryptor.Update([]byte("hello"))
	ciphertext2, _ := encryptor.Final()
	ciphertext := make([]byte, 0, len(ciphertext1)+len(ciphertext2))
	ciphertext = append(ciphertext, ciphertext1...)
	ciphertext = append(ciphertext, ciphertext2...)

	decryptor, _ := gmssl.NewCipherContext("SMS4", key, iv, false)
	plaintext1, _ := decryptor.Update(ciphertext)
	plaintext2, _ := decryptor.Final()
	plaintext := make([]byte, 0, len(plaintext1)+len(plaintext2))
	plaintext = append(plaintext, plaintext1...)
	plaintext = append(plaintext, plaintext2...)

	fmt.Printf("sms4(\"%s\") = %x\n", plaintext, ciphertext)
	fmt.Println()

	/* private key */
	rsa_args := [][2]string{
		{"rsa_keygen_bits", "2048"},
		{"rsa_keygen_pubexp", "65537"},
	}

	rsa, err := gmssl.GeneratePrivateKey("RSA", rsa_args, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	rsa_pem, err := rsa.GetPublicKeyPEM()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(rsa_pem)

	/* Engine */
	eng, _ := gmssl.NewEngineByName(engines[1])
	cmds, _ := eng.GetCommands()
	for _, cmd := range cmds {
		fmt.Print(" " + cmd)
	}
	fmt.Println()

	/* SM2 key pair operations */
	sm2keygenargs := [][2]string{
		{"ec_paramgen_curve", "sm2p256v1"},
		{"ec_param_enc", "named_curve"},
	}
	sm2sk, _ := gmssl.GeneratePrivateKey("EC", sm2keygenargs, nil)
	sm2sktxt, _ := sm2sk.GetText()
	sm2skpem, _ := sm2sk.GetPEM("SMS4", "password")
	sm2pkpem, _ := sm2sk.GetPublicKeyPEM()

	fmt.Println(sm2sktxt)
	fmt.Println(sm2skpem)
	fmt.Println(sm2pkpem)

	sm2pk, _ := gmssl.NewPublicKeyFromPEM(sm2pkpem)
	sm2pktxt, _ := sm2pk.GetText()
	sm2pkpem_, _ := sm2pk.GetPEM()

	fmt.Println(sm2pktxt)
	fmt.Println(sm2pkpem_)

	/* SM2 sign/verification */
	sm2zid, _ := sm2pk.ComputeSM2IDDigest("1234567812345678")
	sm3ctx.Reset()
	sm3ctx.Update(sm2zid)
	sm3ctx.Update([]byte("message"))
	tbs, _ := sm3ctx.Final()

	sig, _ := sm2sk.Sign("sm2sign", tbs, nil)
	fmt.Printf("sm2sign(sm3(\"message\")) = %x\n", sig)

	if ret := sm2pk.Verify("sm2sign", tbs, sig, nil); ret != nil {
		fmt.Printf("sm2 verify failure\n")
	} else {
		fmt.Printf("sm2 verify success\n")
	}

	/* SM2 encrypt */
	sm2msg := "01234567891123456789212345678931234567894123456789512345678961234567897123"
	sm2encalg := "sm2encrypt-with-sm3"
	sm2ciphertext, _ := sm2pk.Encrypt(sm2encalg, []byte(sm2msg), nil)
	sm2plaintext, _ := sm2sk.Decrypt(sm2encalg, sm2ciphertext, nil)
	fmt.Printf("sm2enc(\"%s\") = %x\n", sm2plaintext, sm2ciphertext)
	if sm2msg != string(sm2plaintext) {
		fmt.Println("SM2 encryption/decryption failure")
	}

	/* Certificate */
	certpem := `-----BEGIN CERTIFICATE-----
MIICAjCCAaigAwIBAgIBATAKBggqgRzPVQGDdTBSMQswCQYDVQQGEwJDTjELMAkG
A1UECAwCQkoxCzAJBgNVBAcMAkJKMQwwCgYDVQQKDANQS1UxCzAJBgNVBAsMAkNB
MQ4wDAYDVQQDDAVQS1VDQTAeFw0xNzA2MDEwMDAwMDBaFw0yMDA2MDEwMDAwMDBa
MEYxCzAJBgNVBAYTAkNOMQswCQYDVQQIDAJCSjEMMAoGA1UECgwDUEtVMQswCQYD
VQQLDAJDQTEPMA0GA1UEAwwGYW50c3NzMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0D
QgAEHpXtrYNlwesl7IyPuaHKKHqn4rHBk+tCU0l0T+zuBNMHAOJzKNDbobno6gOI
EQlVfC9q9uk9lO174GJsMLWJJqN7MHkwCQYDVR0TBAIwADAsBglghkgBhvhCAQ0E
HxYdT3BlblNTTCBHZW5lcmF0ZWQgQ2VydGlmaWNhdGUwHQYDVR0OBBYEFJsrRYOA
J8gpNq0KK6yuh/Dv9SjaMB8GA1UdIwQYMBaAFH1Dhf9CqQQYHF/8euzcPROIzn0r
MAoGCCqBHM9VAYN1A0gAMEUCIQCjrQ2nyiPqod/gZdj5X1+WW4fGtyqXvXLL3lOF
31nA/gIgZOpHLnvkyggY9VFfEQVp+8t6kewSfxb4eOImSu+dZcE=
-----END CERTIFICATE-----`

	cert, _ := gmssl.NewCertificateFromPEM(certpem, "")
	subject, _ := cert.GetSubject()
	issuer, _ := cert.GetIssuer()
	serial, _ := cert.GetSerialNumber()
	certpk, _ := cert.GetPublicKey()
	certpktxt, _ := certpk.GetText()
	certtxt, _ := cert.GetText()

	fmt.Println("Certificate:")
	fmt.Printf("  Subject = %s\n", subject)
	fmt.Printf("  Issuer = %s \n", issuer)
	fmt.Printf("  Serial Number = %s\n", serial)
	fmt.Println(certpktxt)
	fmt.Println(certtxt)

	/* SSL */
	hostname := "its.pku.edu.cn"
	ssl, _ := gmssl.NewSSLContext("3.3", "mozilla-cacerts.pem", "")
	conn, _ := ssl.Connect(hostname, "443", "ALL")
	result, _ := conn.GetVerifyResult()
	if result != 0 {
		fmt.Printf("http://%s certificate verify failure\n", hostname)
		return
	}
	peercert, _ := conn.GetPeerCertificate()
	fmt.Println(result)
	peercerttxt, _ := peercert.GetText()
	fmt.Println(peercerttxt)
}
