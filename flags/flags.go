package flags

import (
	"fmt"
	"git.sophuwu.com/gophuwu/parsers"
	"os"
)

type flag struct {
	Name    string
	Short   string
	HelpMsg string
	Type    string
	Default interface{}
	Value   interface{}
}

var flags = make(map[string]flag)
var shortFlags = make(map[byte]string)

// NewFlag creates a new flag with the given name, help message, and default value.
// It returns an error if the flag name is invalid, if the flag already exists,
// or if the default value is of an unsupported type.
// Supported types for defaultValue are: string, int, bool, and float64.
// short must be a single character and unique across all flags.
// set short to an empty string if no short flag is needed.
func NewFlag(name, short, helpMsg string, defaultValue interface{}) error {
	if len(name) == 0 {
		return fmt.Errorf("flag name cannot be empty")
	}
	if _, exists := flags[name]; exists {
		return fmt.Errorf("flag %s already exists", name)
	}
	if defaultValue == nil {
		return fmt.Errorf("default value for flag %s cannot be nil", name)
	}
	var f flag
	f.Name = name
	f.HelpMsg = helpMsg
	if len(short) > 1 {
		return fmt.Errorf("short flag must be a single character")
	}
	if len(short) == 1 {
		shrt := short[0]
		if _, exists := shortFlags[shrt]; exists {
			return fmt.Errorf("short flag %s already exists", short)
		}
		shortFlags[shrt] = name
		f.Short = short
	}
	switch defaultValue.(type) {
	case string:
		f.Type = "string"
		break
	case int:
		f.Type = "int"
		break
	case bool:
		f.Type = "bool"
		break
	case float64:
		f.Type = "float64"
		break
	default:
		return fmt.Errorf("unsupported flag type for %s", name)
	}
	f.Default = defaultValue
	flags[name] = f
	return nil
}

func getFlag(name, t string) (interface{}, error) {
	f, exists := flags[name]
	if !exists {
		return nil, fmt.Errorf("flag %s does not exist", name)
	}
	if f.Type != t {
		return nil, fmt.Errorf("flag %s is not of type bool", name)
	}
	if f.Value == nil {
		return f.Default, nil
	}
	return f.Value, nil
}
func GetBoolFlag(name string) (bool, error) {
	i, err := getFlag(name, "bool")
	if err != nil {
		return false, err
	}
	return i.(bool), nil
}
func GetIntFlag(name string) (int, error) {
	i, err := getFlag(name, "int")
	if err != nil {
		return 0, err
	}
	return i.(int), nil
}
func GetStringFlag(name string) (string, error) {
	i, err := getFlag(name, "string")
	if err != nil {
		return "", err
	}
	return i.(string), nil
}

func GetFloat64Flag(name string) (float64, error) {
	i, err := getFlag(name, "float64")
	if err != nil {
		return 0, err
	}
	return i.(float64), nil
}

func ParseArgs() error {
	if len(os.Args) < 2 {
		return nil
	}
	var v string
	var vv byte
	var i, j int
	var f flag
	var ok bool
	var err error

	shortFlags['h'] = "help"

	var args []string
	for i = 1; i < len(os.Args); i++ {
		v = os.Args[i]
		if (len(v) > 2 && v[0] == '-' && v[1] != '-') || (len(v) == 2 && v[0] == '-') {
			for j = 1; j < len(os.Args[i]); j++ {
				vv = os.Args[i][j]
				v, ok = shortFlags[vv]
				if !ok {
					return fmt.Errorf("unknown short flag: %c", vv)
				}
				args = append(args, "--"+v)
			}
			continue
		}
		args = append(args, v)
	}

	for i = 0; i < len(args); i++ {
		v = args[i]
		if len(v) > 2 && v[0] == '-' && v[1] == '-' {
			v = v[2:]
			if v == "help" {
				PrintHelp()
				os.Exit(0)
			}
			f, ok = flags[v]
			if !ok {
				return fmt.Errorf("unknown flag: %s", v)
			}
			if f.Type == "bool" {
				f.Value = !f.Default.(bool)
				flags[f.Name] = f
				continue
			}
			i++
			v = args[i]
			if i >= len(args) {
				return fmt.Errorf("flag %s requires a value", v)
			}
			switch f.Type {
			case "string":
				f.Value = v
				break
			case "int":
				f.Value, err = parsers.ParseInt(v)
				break
			case "float64":
				f.Value, err = parsers.ParseFloat(v)
				break
			default:
				return fmt.Errorf("unsupported flag type for %s", f.Name)
			}
			if err != nil {
				return fmt.Errorf("error parsing flag %s: %v", f.Name, err)
			}
			flags[f.Name] = f
		}
	}
	return nil
}

func PrintHelp() {
	fmt.Printf("Usage: %s [options]\n", os.Args[0])
	fmt.Println("Options:")
	fmt.Println("  -h --help\n\tShow this help message")
	for _, f := range flags {
		fmt.Printf("  ")
		if len(f.Short) == 1 {
			fmt.Printf("-%s, ", f.Short)
		} else {
			fmt.Printf("    ")
		}
		fmt.Printf("--%s ", f.Name)
		if f.Type == "bool" {
			fmt.Printf("\n")
		} else {
			fmt.Printf("<%s>\n", f.Type)
		}
		fmt.Printf("\t%s ", f.HelpMsg)
		if f.Default != nil {
			fmt.Printf("(default: %v)\n", f.Default)
		} else {
			fmt.Printf("\n")
		}
	}
}
