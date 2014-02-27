package config

import (
	"strings"

	"github.com/robfig/revel"
)

type NestedConfig struct {
	conf *revel.MergedConfig
}

var (
	NestedRevelConfig *NestedConfig
)

func parentSection(mode string) string {
	s := strings.Split(mode, ".")
	if len(s) == 0 {
		return mode
	} else {
		return s[0]
	}
}

func (c *NestedConfig) IntInSection(option string, section string) (int, bool) {
	mode := revel.RunMode
	c.conf.SetSection(section)
	defer c.conf.SetSection(mode)

	// Allow one to reference other variables
	if val, found := c.conf.String(option); found {
		if t, f := c.conf.Int(val); f {
			return t, f
		}
	}

	if val, found := c.conf.Int(option); found {
		return val, found
	}

	if parentSection(section) == section {
		return 0, false
	}

	val, found := c.IntInSection(option, parentSection(section))
	if !found {
		revel.ERROR.Println("Cannnot find value for ", option)
	}
	return val, found
}

func (c *NestedConfig) Int(option string) (int, bool) {
	return c.IntInSection(option, revel.RunMode)
}

func (c *NestedConfig) IntDefaultInSection(option string, section string, deflt int) int {
	if val, found := c.IntInSection(option, section); found {
		return val
	}
	return deflt
}

func (c *NestedConfig) IntDefault(option string, deflt int) int {
	if val, found := c.Int(option); found {
		return val
	}
	return deflt
}

func (c *NestedConfig) StringInSection(option string, section string) (string, bool) {
	mode := revel.RunMode
	c.conf.SetSection(section)
	defer c.conf.SetSection(mode)

	// Allow one to reference other variables
	if val, found := c.conf.String(option); found {
		if t, f := c.conf.String(val); f {
			return t, f
		}
	}

	if val, found := c.conf.String(option); found {
		return val, found
	}

	if parentSection(section) == section {
		return "", false
	}

	val, found := c.StringInSection(option, parentSection(section))
	if !found {
		revel.ERROR.Println("Cannnot find value for ", option)
	}
	return val, found
}

func (c *NestedConfig) String(option string) (string, bool) {
	return c.StringInSection(option, revel.RunMode)
}

func (c *NestedConfig) StringDefaultInSection(option string, section string, deflt string) string {
	if val, found := c.StringInSection(option, section); found {
		return val
	}
	return deflt
}

func (c *NestedConfig) StringDefault(option string, deflt string) string {
	if val, found := c.String(option); found {
		return val
	}
	return deflt
}

func (c *NestedConfig) BoolInSection(option string, section string) (bool, bool) {
	mode := revel.RunMode
	c.conf.SetSection(section)
	defer c.conf.SetSection(mode)

	// Allow one to reference other variables
	if val, found := c.conf.String(option); found {
		if t, f := c.conf.Bool(val); f {
			return t, f
		}
	}

	if val, found := c.conf.Bool(option); found {
		return val, found
	}

	if parentSection(section) == section {
		return false, false
	}

	val, found := c.BoolInSection(option, parentSection(section))
	if !found {
		revel.ERROR.Println("Cannnot find value for ", option)
	}
	return val, found
}

func (c *NestedConfig) Bool(option string) (bool, bool) {
	return c.BoolInSection(option, revel.RunMode)
}

func (c *NestedConfig) BoolDefaultInSection(option string, section string, deflt bool) bool {
	if val, found := c.BoolInSection(option, section); found {
		return val
	}
	return deflt
}

func (c *NestedConfig) BoolDefault(option string, deflt bool) bool {
	if val, found := c.Bool(option); found {
		return val
	}
	return deflt
}

func InitConfigReader() {
	NestedRevelConfig = &NestedConfig{revel.Config}
}
