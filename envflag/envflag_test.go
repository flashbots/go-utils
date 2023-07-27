package envflag_test

import (
	"flag"
	"os"
	"testing"

	"github.com/flashbots/go-utils/envflag"
	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	const name = "bool-var"
	const env = "BOOL_VAR"

	args := make([]string, len(os.Args))
	copy(os.Args, args)
	defer func() {
		os.Args = make([]string, len(args))
		copy(args, os.Args)
	}()

	{ // cli: absent;  env: absent;  default: false
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test"}
		os.Unsetenv(env)
		f := envflag.MustBool(name, false, "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.False(t, *f)
	}
	{ // cli: absent;  env: absent;  default: true
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test"}
		os.Unsetenv(env)
		f := envflag.MustBool(name, true, "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.True(t, *f)
	}
	{ // cli: absent;  env: false;  default: false
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test"}
		t.Setenv(env, "0")
		f := envflag.MustBool(name, false, "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.False(t, *f)
	}
	{ // cli: absent;  env: false;  default: true
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test"}
		t.Setenv(env, "0")
		f := envflag.MustBool(name, true, "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.False(t, *f)
	}
	{ // cli: absent;  env: true;  default: false
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test"}
		t.Setenv(env, "1")
		f := envflag.MustBool(name, false, "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.True(t, *f)
	}
	{ // cli: absent;  env: true;  default: true
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test"}
		t.Setenv(env, "1")
		f := envflag.MustBool(name, true, "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.True(t, *f)
	}

	{ // cli: false;  env: absent;  default: false
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test", "-" + name + "=false"}
		os.Unsetenv(env)
		f := envflag.MustBool(name, false, "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.False(t, *f)
	}
	{ // cli: false;  env: absent;  default: true
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test", "-" + name + "=false"}
		os.Unsetenv(env)
		f := envflag.MustBool(name, true, "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.False(t, *f)
	}
	{ // cli: false;  env: false;  default: false
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test", "-" + name + "=false"}
		t.Setenv(env, "0")
		f := envflag.MustBool(name, false, "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.False(t, *f)
	}
	{ // cli: false;  env: false;  default: true
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test", "-" + name + "=false"}
		t.Setenv(env, "0")
		f := envflag.MustBool(name, true, "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.False(t, *f)
	}
	{ // cli: false;  env: true;  default: false
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test", "-" + name + "=false"}
		t.Setenv(env, "1")
		f := envflag.MustBool(name, false, "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.False(t, *f)
	}
	{ // cli: false;  env: true;  default: true
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test", "-" + name + "=false"}
		t.Setenv(env, "1")
		f := envflag.MustBool(name, true, "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.False(t, *f)
	}

	{ // cli: true;  env: absent;  default: false
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test", "-" + name}
		os.Unsetenv(env)
		f := envflag.MustBool(name, false, "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.True(t, *f)
	}
	{ // cli: true;  env: absent;  default: true
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test", "-" + name}
		os.Unsetenv(env)
		f := envflag.MustBool(name, true, "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.True(t, *f)
	}
	{ // cli: true;  env: false;  default: false
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test", "-" + name}
		t.Setenv(env, "0")
		f := envflag.MustBool(name, false, "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.True(t, *f)
	}
	{ // cli: true;  env: false;  default: true
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test", "-" + name}
		t.Setenv(env, "0")
		f := envflag.MustBool(name, true, "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.True(t, *f)
	}
	{ // cli: true;  env: true;  default: false
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test", "-" + name}
		t.Setenv(env, "1")
		f := envflag.MustBool(name, false, "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.True(t, *f)
	}
	{ // cli: true;  env: true;  default: true
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test", "-" + name}
		t.Setenv(env, "1")
		f := envflag.MustBool(name, true, "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.True(t, *f)
	}
}

func TestInt(t *testing.T) {
	const name = "int-var"
	const env = "INT_VAR"

	args := make([]string, len(os.Args))
	copy(os.Args, args)
	defer func() {
		os.Args = make([]string, len(args))
		copy(args, os.Args)
	}()

	{ // cli: absent;  env: absent;  default: 42
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test"}
		os.Unsetenv(env)
		f := envflag.MustInt(name, 42, "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.Equal(t, 42, *f)
	}
	{ // cli: absent;  env: 42;  default: 0
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test"}
		t.Setenv(env, "42")
		f := envflag.MustInt(name, 0, "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.Equal(t, 42, *f)
	}
	{ // cli: 42;  env: 21;  default: 0
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test", "-" + name, "42"}
		t.Setenv(env, "21")
		f := envflag.MustInt(name, 0, "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.Equal(t, 42, *f)
	}
}

func TestString(t *testing.T) {
	const name = "string-var"
	const env = "STRING_VAR"

	args := make([]string, len(os.Args))
	copy(os.Args, args)
	defer func() {
		os.Args = make([]string, len(args))
		copy(args, os.Args)
	}()

	{ // cli: absent;  env: absent;  default: 42
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test"}
		os.Unsetenv(env)
		f := envflag.String(name, "42", "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.Equal(t, "42", *f)
	}
	{ // cli: absent;  env: 42;  default: 0
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test"}
		t.Setenv(env, "42")
		f := envflag.String(name, "0", "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.Equal(t, "42", *f)
	}
	{ // cli: 42;  env: 21;  default: 0
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		os.Args = []string{"envflag.test", "-" + name, "42"}
		t.Setenv(env, "21")
		f := envflag.String(name, "0", "")
		assert.NotNil(t, f)
		flag.Parse()
		assert.Equal(t, "42", *f)
	}
}
