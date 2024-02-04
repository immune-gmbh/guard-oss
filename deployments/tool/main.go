package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	"github.com/immune-gmbh/guard-oss/deployments/tool/key"
)

var releaseId string = "unknown"

func main() {
	var cmdKeys = &cobra.Command{
		Use:   "keys [key command]",
		Short: "Signing key related commands",
	}

	var cmdNew = &cobra.Command{
		Use:   "new [NAME]",
		Short: "Generate new key pair",
		Long:  ``,
		Args:  cobra.MinimumNArgs(0),
		RunE:  doNew,
	}

	var cmdPrint = &cobra.Command{
		Use:   "print [FILE or -]",
		Short: "Print key information",
		Long:  ``,
		Args:  cobra.MinimumNArgs(0),
		RunE:  doShow,
	}

	var argv0 = "ops"
	if len(os.Args) > 0 && os.Args[0] != "" {
		argv0 = os.Args[0]
	}
	var rootCmd = &cobra.Command{
		Use:   argv0,
		Short: "immune Guard operations tool",
	}
	rootCmd.AddCommand(cmdKeys)
	cmdKeys.AddCommand(cmdNew, cmdPrint)
	rootCmd.Execute()
}

func doShow(cmd *cobra.Command, args []string) error {
	var (
		buf []byte
		err error
	)
	if len(args) > 0 {
		buf, err = os.ReadFile(args[0])
	} else {
		fmt.Fprintln(os.Stderr, "Reading from stdin")
		buf, err = ioutil.ReadAll(os.Stdin)
	}
	if err != nil {
		return err
	}

	pub, priv := tryDecodeKey(buf)
	if priv != nil {
		printPrivateKey(priv)
	} else if pub != nil {
		printPublicKey(pub)
	} else {
		return errors.New("Neither PKCS8 private key nor PKIX public key")
	}
	return nil
}

func doNew(cmd *cobra.Command, args []string) error {
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	pkcs8, err := x509.MarshalECPrivateKey(k)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		err = os.WriteFile(fmt.Sprint(args[0], ".key"), []byte(base64.StdEncoding.EncodeToString(pkcs8)), 0400)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("Private key: %s\n", base64.StdEncoding.EncodeToString(pkcs8))
	}

	pkix, err := x509.MarshalPKIXPublicKey(&k.PublicKey)
	if err != nil {
		return err
	}
	key, err := key.NewKey("", pkix)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		err = os.WriteFile(fmt.Sprint(args[0], ".pub"), []byte(base64.StdEncoding.EncodeToString(pkix)), 0644)
		if err != nil {
			return err
		}
		fmt.Printf("Key with ID %s written to %s.key and %s.pub\n", key.Kid, args[0], args[0])
	} else {
		fmt.Printf("Public key: %s\n", base64.StdEncoding.EncodeToString(pkix))
	}

	return nil
}

func tryDecodeKey(buf []byte) (crypto.PublicKey, crypto.PrivateKey) {
	// try PEM decode first
	if pemBuf, err := pem.Decode(buf); err == nil {
		buf = pemBuf.Bytes
	}

	// try bare base64
	if b64Buf, err := base64.StdEncoding.DecodeString(string(buf)); err == nil {
		buf = b64Buf
	}

	// private key?
	if key, err := x509.ParsePKCS8PrivateKey(buf); err == nil {
		return nil, key
	}

	// public key?
	if pub, err := x509.ParsePKIXPublicKey(buf); err == nil {
		return pub, nil
	}

	// bare ECC key?
	if key, err := x509.ParseECPrivateKey(buf); err == nil {
		return nil, key
	}

	return nil, nil
}

func printPrivateKey(key crypto.PrivateKey) {
	switch key.(type) {
	case *ecdsa.PrivateKey:
		ec := key.(*ecdsa.PrivateKey)

		switch ec.Curve {
		case elliptic.P224():
			panic("Not on NIST P256 but on P224")
		case elliptic.P256():
			if pkcs8, err := x509.MarshalECPrivateKey(ec); err == nil {
				fmt.Printf("Private key (bare EC):    %s\n", base64.StdEncoding.EncodeToString(pkcs8))
				printPublicKey(ec.PublicKey)
			} else {
				panic(err)
			}
		case elliptic.P384():
			panic("Not on NIST P256 but on P384")
		case elliptic.P521():
			panic("Not on NIST P256 but on P521")
		default:
			panic("Not on NIST P256 but on an unknown curve")
		}

	case *rsa.PrivateKey:
		panic("RSA private key")
	default:
		panic("Unknown private key type")
	}
}

func printPublicKey(pub crypto.PublicKey) {
	switch k := pub.(type) {
	case ecdsa.PublicKey:
		printPublicKey(&k)

	case *ecdsa.PublicKey:
		switch k.Curve {
		case elliptic.P224():
			panic("Not on NIST P256 but on P224")
		case elliptic.P256():
			buf, err := x509.MarshalPKIXPublicKey(k)
			if err != nil {
				panic(err)
			}
			key, err := key.NewKey("", buf)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Public key (base64 PKIX): %s\n", base64.StdEncoding.EncodeToString(buf))
			fmt.Printf("Public key ID:            %s\n", key.Kid)
		case elliptic.P384():
			panic("Not on NIST P256 but on P384")
		case elliptic.P521():
			panic("Not on NIST P256 but on P521")
		default:
			panic("Not on NIST P256 but on an unknown curve")
		}

	case *rsa.PublicKey:
		panic("RSA public key")
	default:
		fmt.Printf("%#v\n", pub)
		panic("Unknown public key type")
	}
}
