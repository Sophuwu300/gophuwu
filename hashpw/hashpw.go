package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"git.sophuwu.com/gophuwu/flags"
	"git.sophuwu.com/gophuwu/parsers"
	"golang.org/x/term"
	"hash"
	"io"
	"os"
	"strings"
)

func fatal(err error, msg string) {
	if err != nil {
		println("fatal: " + msg + ": " + err.Error())
		os.Exit(1)
	}
}

func ErrHand(err error) {
	fmt.Printf("Fatal Error: %s\n", err.Error())
	os.Exit(1)
}

func init() {
	newFlag := flags.NewNewFlagWithHandler(ErrHand)
	newFlag("algorithm", "a", "hash algorithm to use, -l to see list", "sha256")
	newFlag("encoder", "e", "text encoder to use, -l to see list", "hex")
	newFlag("list", "l", "list of available algorithms and encoders", false)
	newFlag("output", "o", "write output to file instead of stdout", "stdout")
	newFlag("newline", "n", "add a newline to the end of the output, auto/yes/no", "auto")
	newFlag("show", "s", "do not hide password on manual input", false)
	newFlag("password", "p", "use <string> as password instead of reading stdin", "")
	newFlag("file", "f", "hash a file instead of a password", "")
	err := flags.ParseArgs()
	fatal(err, "could not parse flags")
}

var encoders = map[string]func(src []byte) []byte{
	"base32": func(src []byte) []byte {
		return []byte(base32.StdEncoding.EncodeToString(src))
	},
	"base64": func(src []byte) []byte {
		return []byte(base64.StdEncoding.EncodeToString(src))
	},
	"base64url": func(src []byte) []byte {
		return []byte(base64.URLEncoding.EncodeToString(src))
	},
	"hex": func(src []byte) []byte {
		return []byte(hex.EncodeToString(src))
	},
	"raw": func(src []byte) []byte {
		return src
	},
}
var shortEncoders = map[string]string{
	"32":  "base32",
	"64":  "base64",
	"url": "base64url",
	"16":  "hex",
	"b":   "raw",
}

var autoNewline = true

func getEnc() func(src []byte) []byte {
	encoder, err := flags.GetStringFlag("encoder")
	fatal(err, "could not get encoder flag")
	encoder = strings.ToLower(encoder)
	if s, ok := shortEncoders[encoder]; ok {
		encoder = s
	}
	autoNewline = encoder != "raw" && autoNewline
	if enc, ok := encoders[encoder]; ok {
		return enc
	}
	return nil
}

const algosList = "md5 sha1 sha256 sha256_224 sha512 sha512_224 sha512_256 sha512_384 sha3_224 sha3_256 sha3_384 sha3_512"

var fmtSPad string = "%-5s"

func init() {
	var h func() hash.Hash
	var ok bool
	for algo := range algosListShort {
		if h, ok = algos[algo]; !ok {
			fatal(errors.New("unknown algorithm: "+algo), "invalid algorithm")
		}
		algos[algosListShort[algo]] = h
	}
	var fmtNameLength = 5
	for s := range algos {
		if len(s) > fmtNameLength {
			fmtNameLength = len(s)
		}
	}
	for s := range encoders {
		if len(s) > fmtNameLength {
			fmtNameLength = len(s)
		}
	}
	fmtNameLength += 1
	fmtSPad = fmt.Sprintf("%%-%ds", fmtNameLength)
}

var algosListShort = map[string]string{
	"sha1":       "1",
	"sha512_384": "3",
	"sha512":     "5",
}

var algos = map[string]func() hash.Hash{
	"md5":        func() hash.Hash { return md5.New() },
	"sha1":       func() hash.Hash { return sha1.New() },
	"sha256":     func() hash.Hash { return sha256.New() },
	"sha256_224": func() hash.Hash { return sha256.New224() },
	"sha512":     func() hash.Hash { return sha512.New() },
	"sha512_224": func() hash.Hash { return sha512.New512_224() },
	"sha512_384": func() hash.Hash { return sha512.New384() },
	"sha512_256": func() hash.Hash { return sha512.New512_256() },
	"sha3_224":   func() hash.Hash { return sha3.New224() },
	"sha3_256":   func() hash.Hash { return sha3.New256() },
	"sha3_384":   func() hash.Hash { return sha3.New384() },
	"sha3_512":   func() hash.Hash { return sha3.New512() },
}

func getHashFunc() hash.Hash {
	algorithm, err := flags.GetStringFlag("algorithm")
	fatal(err, "could not get algorithm flag")
	algorithm = strings.ReplaceAll(algorithm, "-", "_")
	if h, ok := algos[algorithm]; ok {
		return h()
	}
	fatal(errors.New("unknown algorithm: "+algorithm), "invalid algorithm")
	return nil
}

func getOutFunc() func(w []byte) error {
	output, err := flags.GetStringFlag("output")
	fatal(err, "could not get output flag")
	if output == "stdout" {
		autoNewline = autoNewline && term.IsTerminal(1)
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

func getInput() []byte {
	var err error
	var s string
	var password []byte
	s, err = flags.GetStringFlag("file")
	fatal(err, "could not get file flag")
	if s != "" {
		password, err = os.ReadFile(s)
		fatal(err, "could not read file")
		if len(password) == 0 {
			fatal(errors.New("file is empty"), "invalid file")
		}
		return password
	}
	s, err = flags.GetStringFlag("password")
	fatal(err, "could not get password flag")
	if s != "" {
		return []byte(s)
	}
	if !term.IsTerminal(0) {
		password, err = io.ReadAll(os.Stdin)
		fatal(err, "could not read from stdin")
		return password
	}

	var show bool
	show, err = flags.GetBoolFlag("show")
	fatal(err, "could not get show flag")
	if show {
		fmt.Print("Enter password to hash: ")
		password, err = parsers.ReadLine(os.Stdin)
		fatal(err, "could not read password")
		return password
	}
	fmt.Print("Enter password to hash: ")
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
	return password
}

func pad(s string, l int) string {
	if len(s) >= l {
		return s
	}
	return fmt.Sprintf(fmt.Sprintf("%%-%ds", l), s)
}

func listAlgos() {
	ok, err := flags.GetBoolFlag("list")
	fatal(err, "could not get list flag")
	if !ok {
		return
	}
	fmt.Println("Available algorithms (short):")
	var short string
	var algo string
	for _, algo = range strings.Split(algosList, " ") {
		if short, ok = algosListShort[algo]; ok {
			fmt.Printf("  - "+fmtSPad+" (%s)\n", algo, short)
			continue
		}
		fmt.Printf("  - "+fmtSPad+"\n", algo)
	}
	fmt.Println("\nAvailable encoders (short):")
	for sh, lg := range shortEncoders {
		fmt.Printf("  - "+fmtSPad+" (%s)\n", lg, sh)
	}
	os.Exit(0)
}

func main() {
	listAlgos()

	var err error
	hashFunc := getHashFunc()
	enc := getEnc()
	out := getOutFunc()

	var line string
	line, err = flags.GetStringFlag("newline")
	fatal(err, "could not get newline flag")
	password := getInput()
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
	if line == "yes" || (line == "auto" && autoNewline) {
		encHash = append(encHash, '\n')
	}
	err = out(encHash)
	fatal(err, "could not write output")

}
