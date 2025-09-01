package flags

import (
	"fmt"
	"os"
	"slices"

	"git.sophuwu.com/gophuwu/parsers"
)

type flag struct {
	Name    string
	Short   string
	Type    string
	Help    []string
	Default interface{}
	Value   interface{}
}

func (f flag) String() string {
	var s string
	if len(f.Short) == 1 {
		s += fmt.Sprintf("-%s, ", f.Short)
	} else {
		s += fmt.Sprintf("    ")
	}
	s += fmt.Sprintf("--%s ", f.Name)
	if f.Type == "bool" {
		s += fmt.Sprintf("\n")
	} else {
		s += fmt.Sprintf("<%s>\n", f.Type)
	}
	if len(f.Help) > 0 {
		s += fmt.Sprintf("\t%s ", f.Help[0])
		if f.Default != nil && !(f.Type == "string" && f.Default.(string) == "") {
			s += fmt.Sprintf("(default: %v)\n", f.Default)
		} else {
			s += fmt.Sprintf("\n")
		}
		for _, v := range f.Help[1:] {
			s += fmt.Sprintf("\t%s\n", v)
		}
	}
	return s
}

var FlagList = []flag{
	{
		Name:    "help",
		Short:   "h",
		Help:    []string{"Show this help message"},
		Type:    "bool",
		Default: false,
	},
}

var FlagMap = map[string]*flag{
	"help": &(FlagList[0]),
}
var shortFlags = map[byte]string{
	'h': "help",
}

func AddHelp(flag string, helpMsg string) error {
	f, ok := FlagMap[flag]
	if !ok {
		return fmt.Errorf("flag %s does not exist", flag)
	}
	f.Help = append(f.Help, helpMsg)
	return nil
}

func NewNewFlagWithHandler(handler func(err error)) func(name, short, helpMsg string, defaultValue interface{}) {
	return func(name, short, helpMsg string, defaultValue interface{}) {
		err := NewFlag(name, short, helpMsg, defaultValue)
		if err != nil {
			handler(err)
		}
	}
}

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
	if _, exists := FlagMap[name]; exists {
		return fmt.Errorf("flag %s already exists", name)
	}
	if defaultValue == nil {
		return fmt.Errorf("default value for flag %s cannot be nil", name)
	}
	if len(short) > 1 {
		return fmt.Errorf("short flag must be a single character, or empty. %s is invalid", short)
	}
	var shrt *byte = nil
	if len(short) == 1 {
		b := short[0]
		shrt = &b
		if _, exists := shortFlags[b]; exists {
			return fmt.Errorf("short flag %s already exists", short)
		}
	}
	t := fmt.Sprintf("%T", defaultValue)
	if !slices.Contains([]string{"string", "int", "bool", "float64"}, t) {
		return fmt.Errorf("unsupported type %T in default value. Supported types are: string, int, bool, float64", defaultValue)
	}
	FlagList = append(FlagList, flag{
		Name:    name,
		Short:   short,
		Help:    []string{helpMsg},
		Type:    t,
		Default: defaultValue,
	})
	if shrt != nil {
		shortFlags[*shrt] = name
	}
	FlagMap[name] = &FlagList[len(FlagList)-1]
	return nil
}

func getFlag(name, t string) (interface{}, error) {
	f, exists := FlagMap[name]
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

var otherArgs []string

func OtherArgs() []string {
	return otherArgs
}

func ParseArgs() error {
	if len(os.Args) < 2 {
		return nil
	}
	var v string
	var vv byte
	var i, j int
	var f *flag
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
			f, ok = FlagMap[v]
			if !ok {
				return fmt.Errorf("unknown flag: %s", v)
			}
			if f.Type == "bool" {
				f.Value = !f.Default.(bool)
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
		} else {
			otherArgs = append(otherArgs, args[i])
		}
	}
	if ok, err = GetBoolFlag("help"); err != nil {
		return fmt.Errorf("error getting help flag: %v", err)
	} else if ok {
		PrintHelp()
		os.Exit(0)
	}
	return nil
}

func PrintHelp() {
	fmt.Printf("Usage: %s [options]\n", os.Args[0])
	fmt.Println("Options:")
	for _, f := range FlagList {
		fmt.Print(f.String())
	}
}
