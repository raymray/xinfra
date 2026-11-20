package tagspike

import (
	"reflect"
	"strings"
	"testing"

	"github.com/go-viper/mapstructure/v2"
	"gopkg.in/yaml.v3"
)

// Inner 是内嵌配置段，字段必须导出两个解码器才能处理。
type Inner struct {
	DSN string `yaml:"dsn" mapstructure:"dsn"`
	Max int    `yaml:"max_open_conns" mapstructure:"max_open_conns"`
}

type Item struct {
	Plugin string `yaml:"plugin" mapstructure:"plugin"`
	Inner  `yaml:",inline" mapstructure:",squash"`
}

const doc = `
master:
  plugin: xdb
  dsn: "user:pw@tcp(127.0.0.1:3306)/app"
  max_open_conns: 128
`

func decodeYAMLStrict(s string, out any) error {
	dec := yaml.NewDecoder(strings.NewReader(s))
	dec.KnownFields(true)
	return dec.Decode(out)
}

func TestYamlStrictDecodesUnderscoreKeys(t *testing.T) {
	var got map[string]Item
	if err := decodeYAMLStrict(doc, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	m := got["master"]
	if m.Max != 128 {
		t.Fatalf("Max = %d, want 128 (underscore key dropped?)", m.Max)
	}
	if m.DSN == "" {
		t.Fatal("DSN empty")
	}
	if m.Plugin != "xdb" {
		t.Fatalf("Plugin = %q", m.Plugin)
	}
}

func TestYamlStrictRejectsUnknownKey(t *testing.T) {
	bad := strings.Replace(doc, "max_open_conns", "max_open_conn", 1)
	var got map[string]Item
	err := decodeYAMLStrict(bad, &got)
	if err == nil {
		t.Fatal("expected unknown key error, got nil")
	}
	if !strings.Contains(err.Error(), "max_open_conn") {
		t.Fatalf("error should name the key, got: %v", err)
	}
}

func TestMapstructureTagDecodesSameValues(t *testing.T) {
	var viaYAML map[string]Item
	if err := decodeYAMLStrict(doc, &viaYAML); err != nil {
		t.Fatalf("yaml decode: %v", err)
	}

	var raw map[string]any
	if err := yaml.Unmarshal([]byte(doc), &raw); err != nil {
		t.Fatalf("yaml to any: %v", err)
	}

	var viaMS map[string]Item
	if err := mapstructure.Decode(raw, &viaMS); err != nil {
		t.Fatalf("mapstructure decode: %v", err)
	}

	if !reflect.DeepEqual(viaYAML, viaMS) {
		t.Fatalf("decoders disagree:\n yaml=%+v\n ms  =%+v", viaYAML, viaMS)
	}
}
