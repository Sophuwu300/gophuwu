package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"git.sophuwu.com/gophuwu/flags"
	"git.sophuwu.com/gophuwu/parsers"
	"golang.org/x/term"
	"hash"
	"os"
)

func fatal(err error, msg string) {
	if err != nil {
		println("fatal: " + msg + ": " + err.Error())
		os.Exit(1)
	}
}

func init() {
	err := flags.NewFlag("algorithm", "a", "hash with sha1, sha256 or sha512", "sha256")
	fatal(err, "could not add algorithm flag")
	err = flags.NewFlag("encoder", "e", "encode with base64, base64url, hex or raw", "hex")
	fatal(err, "could not add encoder flag")
	err = flags.NewFlag("output", "o", "output file to write the hashed password to", "stdout")
	fatal(err, "could not add output flag")
	err = flags.NewFlag("newline", "n", "add a newline to the end of the output", false)
	fatal(err, "could not add show flag")
	err = flags.NewFlag("show", "s", "show input when typing", false)
	fatal(err, "could not add show flag")
	err = flags.ParseArgs()
	fatal(err, "could not parse flags")
}

func getEnc() func(src []byte) []byte {
	encoder, err := flags.GetStringFlag("encoder")
	fatal(err, "could not get encoder flag")
	switch encoder {
	case "base64":
		return func(src []byte) []byte {
			return []byte(base64.StdEncoding.EncodeToString(src))
		}
	case "base64url":
		return func(src []byte) []byte {
			return []byte(base64.URLEncoding.EncodeToString(src))
		}
	case "hex":
		return func(src []byte) []byte {
			return []byte(hex.EncodeToString(src))
		}
	case "raw":
		return func(src []byte) []byte {
			return src
		}
	default:
		fatal(errors.New("unknown encoder: "+encoder), "invalid encoder")
	}
	return nil
}

func getHashFunc() hash.Hash {
	algorithm, err := flags.GetStringFlag("algorithm")
	fatal(err, "could not get algorithm flag")
	switch algorithm {
	case "sha1":
		return sha1.New()
	case "sha256":
		return sha256.New()
	case "sha512":
		return sha512.New()
	default:
		fatal(errors.New("unknown algorithm: "+algorithm), "invalid algorithm")
	}
	return nil
}

func getOutFunc() func(w []byte) error {
	output, err := flags.GetStringFlag("output")
	fatal(err, "could not get output flag")
	if output == "stdout" {
		return func(w []byte) error {
			_, e := os.Stdout.Write(w)
			return e
		}
	}
	if output != "" {
		_, err = os.Stat(output)
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Printf("output file already exists, do you want to overwrite it? (y/n) ")
			line, _ := parsers.ReadLineString(os.Stdin)
			if line == "y" {
				fmt.Printf("\roverwriting: %s\n", output)
			}
		}
		return func(w []byte) error {
			f, e := os.Create(output)
			if e != nil {
				return e
			}
			_, e = f.Write(w)
			f.Close()
			return e
		}
	}
	fatal(errors.New("invalid output: "+output), "invalid output")
	return nil
}

func main() {
	var err error

	hashFunc := getHashFunc()
	enc := getEnc()
	out := getOutFunc()

	var line bool
	line, err = flags.GetBoolFlag("newline")
	fatal(err, "could not get newline flag")
	var show bool
	show, err = flags.GetBoolFlag("show")
	fatal(err, "could not get show flag")
	var password []byte

	fmt.Print("Enter password to hash: ")
	if show {
		password, err = parsers.ReadLine(os.Stdin)
		fatal(err, "could not read password")
	} else {
		password, err = term.ReadPassword(0)
		fmt.Println()
		fatal(err, "could not read password")
		fmt.Print("Enter password again: ")
		var pas []byte
		pas, err = term.ReadPassword(0)
		fmt.Println()
		fatal(err, "could not read password")
		if string(pas) != string(password) {
			fmt.Println("Passwords do not match, please try again.")
			os.Exit(1)
		}
	}
	if len(password) == 0 {
		fatal(errors.New("password cannot be empty"), "invalid password")
	}

	var n int
	n, err = hashFunc.Write(password)
	fatal(err, "could not write password to hash function")
	if n != len(password) {
		fatal(errors.New("could not write all bytes to hash function"), "invalid password")
	}
	encHash := hashFunc.Sum(nil)
	encHash = enc(encHash)
	if line {
		encHash = append(encHash, '\n')
	}
	err = out(encHash)
	fatal(err, "could not write output")

}
