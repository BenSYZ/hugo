// Copyright 2019 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package markup_config

import (
	"github.com/gohugoio/hugo/common/maps"
	"github.com/gohugoio/hugo/config"
	"github.com/gohugoio/hugo/docshelper"
	"github.com/gohugoio/hugo/markup/asciidocext/asciidocext_config"
	"github.com/gohugoio/hugo/markup/blackfriday/blackfriday_config"
	"github.com/gohugoio/hugo/markup/goldmark/goldmark_config"
	"github.com/gohugoio/hugo/markup/highlight"
	"github.com/gohugoio/hugo/markup/pandoc/pandoc_config"
	"github.com/gohugoio/hugo/markup/tableofcontents"
	"github.com/gohugoio/hugo/parser"
	"github.com/mitchellh/mapstructure"
)

type Config struct {
	// Default markdown handler for md/markdown extensions.
	// Default is "goldmark".
	// Before Hugo 0.60 this was "blackfriday".
	DefaultMarkdownHandler string

	Highlight       highlight.Config
	TableOfContents tableofcontents.Config

	// Content renderers
	Goldmark    goldmark_config.Config
	BlackFriday blackfriday_config.Config

	AsciidocExt asciidocext_config.Config
	Pandoc      pandoc_config.Config
}

func Decode(cfg config.Provider) (conf Config, err error) {
	conf = Default

	m := cfg.GetStringMap("markup")
	if m == nil {
		return
	}
	normalizeConfig(m)

	err = mapstructure.WeakDecode(m, &conf)
	if err != nil {
		return
	}

	if err = applyLegacyConfig(cfg, &conf); err != nil {
		return
	}

	if err = highlight.ApplyLegacyConfig(cfg, &conf.Highlight); err != nil {
		return
	}

	return
}

func normalizeConfig(m map[string]interface{}) {
	v, err := maps.GetNestedParam("goldmark.parser", ".", m)
	if err != nil {
		return
	}
	vm := maps.ToStringMap(v)
	// Changed from a bool in 0.81.0
	if vv, found := vm["attribute"]; found {
		if vvb, ok := vv.(bool); ok {
			vm["attribute"] = goldmark_config.ParserAttribute{
				Title: vvb,
			}
		}
	}
}

func applyLegacyConfig(cfg config.Provider, conf *Config) error {
	if bm := cfg.GetStringMap("blackfriday"); bm != nil {
		// Legacy top level blackfriday config.
		err := mapstructure.WeakDecode(bm, &conf.BlackFriday)
		if err != nil {
			return err
		}
	}

	if conf.BlackFriday.FootnoteAnchorPrefix == "" {
		conf.BlackFriday.FootnoteAnchorPrefix = cfg.GetString("footnoteAnchorPrefix")
	}

	if conf.BlackFriday.FootnoteReturnLinkContents == "" {
		conf.BlackFriday.FootnoteReturnLinkContents = cfg.GetString("footnoteReturnLinkContents")
	}

	return nil
}

var Default = Config{
	DefaultMarkdownHandler: "goldmark",

	TableOfContents: tableofcontents.DefaultConfig,
	Highlight:       highlight.DefaultConfig,

	Goldmark:    goldmark_config.Default,
	BlackFriday: blackfriday_config.Default,

	AsciidocExt: asciidocext_config.Default,
	Pandoc:      pandoc_config.Default,
}

func init() {
	docsProvider := func() docshelper.DocProvider {
		return docshelper.DocProvider{"config": map[string]interface{}{"markup": parser.LowerCaseCamelJSONMarshaller{Value: Default}}}
	}
	docshelper.AddDocProviderFunc(docsProvider)
}
