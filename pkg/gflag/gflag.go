// SPDX-License-Identifier: CC0-1.0

// Package gflag provides command-line flag parsing with support for both
// short (-x) and long (--flag) flag formats, similar to GNU-style flags.
//
// This package extends the functionality of Go's standard flag package
// to support POSIX-style short flags and GNU-style long flags.
//
// Automatic Help Flag:
// The package automatically adds a --help (-h) flag to all FlagSets unless
// a help flag already exists. When --help or -h is used, it displays usage
// information and exits (or handles based on ErrorHandling setting).
//
// Usage (Traditional API):
//
//	import "github.com/nzions/sharedgolibs/pkg/gflag"
//
//	var verbose = gflag.BoolP("verbose", "v", false, "enable verbose output")
//	var port = gflag.IntP("port", "p", 8080, "server port")
//	var name = gflag.StringP("name", "n", "default", "server name")
//
//	func main() {
//	    gflag.Parse()
//	    if *verbose {
//	        fmt.Println("Verbose mode enabled")
//	    }
//	    fmt.Printf("Server %s listening on port %d\n", *name, *port)
//	}
//
// Usage (Traditional API with *Var functions):
//
//	import "github.com/nzions/sharedgolibs/pkg/gflag"
//
//	var verbose bool
//	var port int
//	var name string
//
//	func init() {
//	    gflag.BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
//	    gflag.IntVarP(&port, "port", "p", 8080, "server port")
//	    gflag.StringVarP(&name, "name", "n", "default", "server name")
//	}
//
//	func main() {
//	    gflag.Parse()
//	    if verbose {
//	        fmt.Println("Verbose mode enabled")
//	    }
//	    fmt.Printf("Server %s listening on port %d\n", name, port)
//	}
//
// Usage (Modern API):
//
//	import "github.com/nzions/sharedgolibs/pkg/gflag"
//
//	func main() {
//	    flags := gflag.New()
//	    flags.AddBool("verbose", "v", false, "enable verbose output")
//	    flags.AddInt("port", "p", 8080, "server port")
//	    flags.AddString("name", "n", "default", "server name")
//
//	    flags.Parse()
//
//	    if flags.GetBool("verbose") {
//	        fmt.Println("Verbose mode enabled")
//	    }
//	    fmt.Printf("Server %s listening on port %d\n",
//	        flags.GetString("name"), flags.GetInt("port"))
//	}
//
// API Functions:
//
// The package provides both Type and TypeP variants for all flag functions:
//   - Type functions: String, Bool, Int, StringVar, BoolVar, IntVar
//   - TypeP functions: StringP, BoolP, IntP, StringVarP, BoolVarP, IntVarP
//
// The P variants accept a short name parameter, while the non-P variants
// only accept the long name.
//
// Supports the following flag formats:
//   - Short flags: -v, -p 8080, -n name
//   - Long flags: --verbose, --port=8080, --name=name
//   - Combined short flags: -vp 8080 (equivalent to -v -p 8080)
//   - Help flags: --help, -h (automatically added)
package gflag

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Version is the current version of the gflag package
const Version = "1.3.0"

// Value represents the interface to the dynamic value stored in a flag.
type Value interface {
	String() string
	Set(string) error
}

// Flag represents a single flag.
type Flag struct {
	Name      string // long name of flag
	ShortName string // short name of flag (single character)
	Usage     string // help message
	Value     Value  // value as set
	DefValue  string // default value (as text); for usage message
}

// FlagSet represents a set of defined flags.
type FlagSet struct {
	name          string
	parsed        bool
	args          []string // arguments after flags
	flags         map[string]*Flag
	shortMap      map[string]*Flag // maps short names to flags
	usage         func()
	errorHandling ErrorHandling
}

// CommandLine is the default set of command-line flags, parsed from os.Args.
var CommandLine = NewFlagSet(os.Args[0], ExitOnError)

// ErrorHandling defines how FlagSet.Parse behaves if the parse fails.
type ErrorHandling int

const (
	ContinueOnError ErrorHandling = iota // Return a descriptive error.
	ExitOnError                          // Call os.Exit(2) or for -h/-help Exit(0).
	PanicOnError                         // Call panic with a descriptive error.
)

// NewFlagSet returns a new, empty flag set with the specified name and
// error handling property.
func NewFlagSet(name string, errorHandling ErrorHandling) *FlagSet {
	f := &FlagSet{
		name:          name,
		flags:         make(map[string]*Flag),
		shortMap:      make(map[string]*Flag),
		errorHandling: errorHandling,
	}
	f.usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", f.name)
		f.PrintDefaults()
	}

	// Automatically add help flags if they don't already exist
	f.addHelpFlagIfNotExists()

	return f
}

// addHelpFlagIfNotExists automatically adds help flags if they don't already exist
func (f *FlagSet) addHelpFlagIfNotExists() {
	// Check if help flag already exists
	if _, exists := f.flags["help"]; !exists {
		// Create a new bool variable for the help flag
		helpPtr := new(bool)
		f.BoolVar(helpPtr, "help", "h", false, "show help message")
	}
}

// boolValue implements Value for bool flags.
type boolValue bool

func newBoolValue(val bool, p *bool) *boolValue {
	*p = val
	return (*boolValue)(p)
}

func (b *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	*b = boolValue(v)
	return nil
}

func (b *boolValue) String() string {
	return strconv.FormatBool(bool(*b))
}

// stringValue implements Value for string flags.
type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) String() string {
	return string(*s)
}

// intValue implements Value for int flags.
type intValue int

func newIntValue(val int, p *int) *intValue {
	*p = val
	return (*intValue)(p)
}

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		return err
	}
	*i = intValue(v)
	return nil
}

func (i *intValue) String() string {
	return strconv.Itoa(int(*i))
}

// Var defines a flag with the specified name, short name, and usage string.
// The type and value of the flag are represented by the first argument, of type Value.
func (f *FlagSet) Var(value Value, name, shortName, usage string) {
	flag := &Flag{
		Name:      name,
		ShortName: shortName,
		Usage:     usage,
		Value:     value,
		DefValue:  value.String(),
	}
	f.flags[name] = flag
	if shortName != "" {
		f.shortMap[shortName] = flag
	}
}

// String defines a string flag with specified name, short name, default value, and usage string.
func (f *FlagSet) String(name, shortName, value, usage string) *string {
	p := new(string)
	f.StringVar(p, name, shortName, value, usage)
	return p
}

// StringVar defines a string flag with specified name, short name, default value, and usage string.
func (f *FlagSet) StringVar(p *string, name, shortName, value, usage string) {
	f.Var(newStringValue(value, p), name, shortName, usage)
}

// Bool defines a bool flag with specified name, short name, default value, and usage string.
func (f *FlagSet) Bool(name, shortName string, value bool, usage string) *bool {
	p := new(bool)
	f.BoolVar(p, name, shortName, value, usage)
	return p
}

// BoolVar defines a bool flag with specified name, short name, default value, and usage string.
func (f *FlagSet) BoolVar(p *bool, name, shortName string, value bool, usage string) {
	f.Var(newBoolValue(value, p), name, shortName, usage)
}

// Int defines an int flag with specified name, short name, default value, and usage string.
func (f *FlagSet) Int(name, shortName string, value int, usage string) *int {
	p := new(int)
	f.IntVar(p, name, shortName, value, usage)
	return p
}

// IntVar defines an int flag with specified name, short name, default value, and usage string.
func (f *FlagSet) IntVar(p *int, name, shortName string, value int, usage string) {
	f.Var(newIntValue(value, p), name, shortName, usage)
}

// Parse parses flag definitions from the argument list, which should not
// include the command name.
func (f *FlagSet) Parse(arguments []string) error {
	f.parsed = true
	f.args = nil

	for i := 0; i < len(arguments); i++ {
		arg := arguments[i]

		if arg == "--" {
			// Everything after -- is not a flag
			f.args = append(f.args, arguments[i+1:]...)
			break
		}

		if !strings.HasPrefix(arg, "-") || arg == "-" {
			// Not a flag (or single dash), add to args
			f.args = append(f.args, arg)
			continue
		}

		var err error
		if strings.HasPrefix(arg, "--") {
			// Long flag
			err = f.parseLongFlag(arg[2:], arguments, &i)
		} else {
			// Short flag(s)
			err = f.parseShortFlag(arg[1:], arguments, &i)
		}

		if err != nil {
			return f.handleError(err)
		}
	}

	return nil
}

// handleError handles errors based on the ErrorHandling setting
func (f *FlagSet) handleError(err error) error {
	switch f.errorHandling {
	case ExitOnError:
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
		return nil // This will never be reached
	case PanicOnError:
		panic(err)
	case ContinueOnError:
		return err
	default:
		return err
	}
}

// parseLongFlag handles --flag or --flag=value format
func (f *FlagSet) parseLongFlag(flagStr string, arguments []string, i *int) error {
	var name, value string
	var hasValue bool

	if idx := strings.Index(flagStr, "="); idx != -1 {
		name = flagStr[:idx]
		value = flagStr[idx+1:]
		hasValue = true
	} else {
		name = flagStr
	}

	flag, exists := f.flags[name]
	if !exists {
		return fmt.Errorf("flag provided but not defined: -%s", name)
	}

	// Special handling for help flag
	if name == "help" {
		f.showHelpAndExit()
		return nil
	}

	// Special handling for bool flags
	if _, isBool := flag.Value.(*boolValue); isBool {
		if hasValue {
			return flag.Value.Set(value)
		} else {
			return flag.Value.Set("true")
		}
	}

	// Non-bool flags need a value
	if hasValue {
		return flag.Value.Set(value)
	}

	// Look for value in next argument
	if *i+1 >= len(arguments) {
		return fmt.Errorf("flag needs an argument: -%s", name)
	}

	*i++
	return flag.Value.Set(arguments[*i])
}

// parseShortFlag handles -f or -f value or combined -abc format
func (f *FlagSet) parseShortFlag(flagStr string, arguments []string, i *int) error {
	for j, char := range flagStr {
		shortName := string(char)
		flag, exists := f.shortMap[shortName]
		if !exists {
			return fmt.Errorf("flag provided but not defined: -%s", shortName)
		}

		// Special handling for help flag (short form)
		if shortName == "h" {
			if helpFlag, helpExists := f.flags["help"]; helpExists && helpFlag.ShortName == "h" {
				f.showHelpAndExit()
				return nil
			}
		}

		// Special handling for bool flags
		if _, isBool := flag.Value.(*boolValue); isBool {
			err := flag.Value.Set("true")
			if err != nil {
				return err
			}
			continue
		}

		// Non-bool flag needs a value
		if j < len(flagStr)-1 {
			// Value is the rest of the flag string
			return flag.Value.Set(flagStr[j+1:])
		}

		// Look for value in next argument
		if *i+1 >= len(arguments) {
			return fmt.Errorf("flag needs an argument: -%s", shortName)
		}

		*i++
		return flag.Value.Set(arguments[*i])
	}

	return nil
}

// showHelpAndExit displays help message and exits based on error handling
func (f *FlagSet) showHelpAndExit() {
	f.usage()
	switch f.errorHandling {
	case ExitOnError:
		os.Exit(0)
	case PanicOnError:
		panic("help requested")
	case ContinueOnError:
		// Do nothing, just return
	}
}

// Args returns the non-flag arguments.
func (f *FlagSet) Args() []string {
	return f.args
}

// NArg returns the number of remaining non-flag arguments.
func (f *FlagSet) NArg() int {
	return len(f.args)
}

// Arg returns the i'th argument. Arg(0) is the first remaining argument
// after flags have been processed.
func (f *FlagSet) Arg(i int) string {
	if i < 0 || i >= len(f.args) {
		return ""
	}
	return f.args[i]
}

// PrintDefaults prints to standard error the default values of all defined flags.
func (f *FlagSet) PrintDefaults() {
	for _, flag := range f.flags {
		format := "  -%s"
		if flag.ShortName != "" {
			format = "  -%s, --%s"
			fmt.Fprintf(os.Stderr, format, flag.ShortName, flag.Name)
		} else {
			fmt.Fprintf(os.Stderr, "      --%s", flag.Name)
		}

		if flag.DefValue != "" && flag.DefValue != "false" {
			fmt.Fprintf(os.Stderr, " (default %q)", flag.DefValue)
		}
		fmt.Fprintf(os.Stderr, "\n        %s\n", flag.Usage)
	}
}

// Modern API methods for cleaner flag handling

// AddBool adds a boolean flag with specified name, short name, default value, and usage string.
func (f *FlagSet) AddBool(name, shortName string, value bool, usage string) {
	f.BoolVar(new(bool), name, shortName, value, usage)
}

// AddString adds a string flag with specified name, short name, default value, and usage string.
func (f *FlagSet) AddString(name, shortName, value, usage string) {
	f.StringVar(new(string), name, shortName, value, usage)
}

// AddInt adds an int flag with specified name, short name, default value, and usage string.
func (f *FlagSet) AddInt(name, shortName string, value int, usage string) {
	f.IntVar(new(int), name, shortName, value, usage)
}

// GetBool returns the value of the named bool flag.
func (f *FlagSet) GetBool(name string) bool {
	flag, exists := f.flags[name]
	if !exists {
		return false
	}
	if boolVal, ok := flag.Value.(*boolValue); ok {
		return bool(*boolVal)
	}
	return false
}

// GetString returns the value of the named string flag.
func (f *FlagSet) GetString(name string) string {
	flag, exists := f.flags[name]
	if !exists {
		return ""
	}
	if stringVal, ok := flag.Value.(*stringValue); ok {
		return string(*stringVal)
	}
	return ""
}

// GetInt returns the value of the named int flag.
func (f *FlagSet) GetInt(name string) int {
	flag, exists := f.flags[name]
	if !exists {
		return 0
	}
	if intVal, ok := flag.Value.(*intValue); ok {
		return int(*intVal)
	}
	return 0
}

// Set sets the value of the named flag.
func (f *FlagSet) Set(name, value string) error {
	flag, exists := f.flags[name]
	if !exists {
		return fmt.Errorf("flag provided but not defined: %s", name)
	}
	return flag.Value.Set(value)
}

// IsSet returns true if the flag was explicitly set during parsing.
func (f *FlagSet) IsSet(name string) bool {
	flag, exists := f.flags[name]
	if !exists {
		return false
	}
	// A flag is considered set if its current value differs from the default
	return flag.Value.String() != flag.DefValue
}

// GetFlag returns the Flag struct for the named flag, or nil if not found.
func (f *FlagSet) GetFlag(name string) *Flag {
	return f.flags[name]
}

// AllFlags returns a map of all flags keyed by their long names.
func (f *FlagSet) AllFlags() map[string]*Flag {
	result := make(map[string]*Flag)
	for name, flag := range f.flags {
		result[name] = flag
	}
	return result
}

// FlagNames returns a slice of all flag names.
func (f *FlagSet) FlagNames() []string {
	names := make([]string, 0, len(f.flags))
	for name := range f.flags {
		names = append(names, name)
	}
	return names
}

// New creates a new FlagSet with convenient defaults for typical usage.
func New() *FlagSet {
	return NewFlagSet("app", ExitOnError)
}

// Package level convenience functions that operate on CommandLine

// StringP defines a string flag with specified name, short name, default value, and usage string.
func StringP(name, shortName, value, usage string) *string {
	return CommandLine.String(name, shortName, value, usage)
}

// String defines a string flag with specified name, default value, and usage string.
func String(name, value, usage string) *string {
	return CommandLine.String(name, "", value, usage)
}

// BoolP defines a bool flag with specified name, short name, default value, and usage string.
func BoolP(name, shortName string, value bool, usage string) *bool {
	return CommandLine.Bool(name, shortName, value, usage)
}

// Bool defines a bool flag with specified name, default value, and usage string.
func Bool(name string, value bool, usage string) *bool {
	return CommandLine.Bool(name, "", value, usage)
}

// IntP defines an int flag with specified name, short name, default value, and usage string.
func IntP(name, shortName string, value int, usage string) *int {
	return CommandLine.Int(name, shortName, value, usage)
}

// Int defines an int flag with specified name, default value, and usage string.
func Int(name string, value int, usage string) *int {
	return CommandLine.Int(name, "", value, usage)
}

// Var defines a flag with the specified name and usage string.
// The type and value of the flag are represented by the first argument, of type Value.
func Var(value Value, name, usage string) {
	CommandLine.Var(value, name, "", usage)
}

// VarP defines a flag with the specified name, short name, and usage string.
// The type and value of the flag are represented by the first argument, of type Value.
func VarP(value Value, name, shortName, usage string) {
	CommandLine.Var(value, name, shortName, usage)
}

// StringVar defines a string flag with specified name, default value, and usage string.
func StringVar(p *string, name, value, usage string) {
	CommandLine.StringVar(p, name, "", value, usage)
}

// StringVarP defines a string flag with specified name, short name, default value, and usage string.
func StringVarP(p *string, name, shortName, value, usage string) {
	CommandLine.StringVar(p, name, shortName, value, usage)
}

// BoolVar defines a bool flag with specified name, default value, and usage string.
func BoolVar(p *bool, name string, value bool, usage string) {
	CommandLine.BoolVar(p, name, "", value, usage)
}

// BoolVarP defines a bool flag with specified name, short name, default value, and usage string.
func BoolVarP(p *bool, name, shortName string, value bool, usage string) {
	CommandLine.BoolVar(p, name, shortName, value, usage)
}

// IntVar defines an int flag with specified name, default value, and usage string.
func IntVar(p *int, name string, value int, usage string) {
	CommandLine.IntVar(p, name, "", value, usage)
}

// IntVarP defines an int flag with specified name, short name, default value, and usage string.
func IntVarP(p *int, name, shortName string, value int, usage string) {
	CommandLine.IntVar(p, name, shortName, value, usage)
}

// Parse parses the command-line flags from os.Args[1:].
func Parse() {
	CommandLine.Parse(os.Args[1:])
}

// Args returns the non-flag command-line arguments.
func Args() []string {
	return CommandLine.Args()
}

// NArg returns the number of remaining non-flag arguments.
func NArg() int {
	return CommandLine.NArg()
}

// Arg returns the i'th command-line argument.
func Arg(i int) string {
	return CommandLine.Arg(i)
}

// Modern API package-level convenience functions

// AddBool adds a bool flag to the default CommandLine flagset.
func AddBool(name, shortName string, value bool, usage string) {
	CommandLine.AddBool(name, shortName, value, usage)
}

// AddString adds a string flag to the default CommandLine flagset.
func AddString(name, shortName, value, usage string) {
	CommandLine.AddString(name, shortName, value, usage)
}

// AddInt adds an int flag to the default CommandLine flagset.
func AddInt(name, shortName string, value int, usage string) {
	CommandLine.AddInt(name, shortName, value, usage)
}

// GetBool returns the value of the named bool flag from CommandLine.
func GetBool(name string) bool {
	return CommandLine.GetBool(name)
}

// GetString returns the value of the named string flag from CommandLine.
func GetString(name string) string {
	return CommandLine.GetString(name)
}

// GetInt returns the value of the named int flag from CommandLine.
func GetInt(name string) int {
	return CommandLine.GetInt(name)
}

// Set sets the value of the named flag in CommandLine.
func Set(name, value string) error {
	return CommandLine.Set(name, value)
}

// IsSet returns true if the flag was explicitly set in CommandLine.
func IsSet(name string) bool {
	return CommandLine.IsSet(name)
}
