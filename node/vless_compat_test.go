package node

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"testing"
)

func TestVLESSMihomoSubscription(t *testing.T) {
	dir := t.TempDir()
	template := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(template, []byte("proxies: []\nproxy-groups:\n  - name: Netflix\n    type: select\n    proxies: [DIRECT]\n  - name: Auto\n    type: url-test\n    url: http://example.com\n    interval: 300\n"), 0600); err != nil {
		t.Fatal(err)
	}
	links := []string{
		"vless://00000000-0000-4000-8000-000000000000@example.com:443?security=reality&fp=chrome&pbk=public&sid=ab&type=tcp&encryption=mlkem768x25519plus.native.0rtt.example#Netflix",
		"vless://00000000-0000-4000-8000-000000000001@example.com:443?security=reality&type=tcp&support-x25519mlkem768=false#old",
		"vless://00000000-0000-4000-8000-000000000002@example.com:443?security=tls&type=xhttp&path=%2Fpath&host=front.example&mode=packet-up#xhttp",
	}
	output, err := EncodeClash(links, SqlConfig{Clash: template})
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Proxies []map[string]interface{} `yaml:"proxies"`
		Groups  []map[string]interface{} `yaml:"proxy-groups"`
	}
	if err := yaml.Unmarshal(output, &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Proxies) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(result.Proxies))
	}
	first := result.Proxies[0]
	if first["name"] != "Netflix (2)" {
		t.Errorf("group collision not resolved: %v", first["name"])
	}
	if first["encryption"] != "mlkem768x25519plus.native.0rtt.example" {
		t.Errorf("missing encryption: %v", first["encryption"])
	}
	reality := first["reality-opts"].(map[string]interface{})
	if reality["support-x25519mlkem768"] != true {
		t.Error("missing hybrid key share")
	}
	if _, ok := first["grpc-opts"]; ok {
		t.Error("unexpected grpc settings for TCP")
	}
	old := result.Proxies[1]["reality-opts"].(map[string]interface{})
	if old["support-x25519mlkem768"] != false {
		t.Error("explicit legacy hint ignored")
	}
	xhttp := result.Proxies[2]["xhttp-opts"].(map[string]interface{})
	if xhttp["host"] != "front.example" || xhttp["path"] != "/path" || xhttp["mode"] != "packet-up" {
		t.Errorf("lost XHTTP options: %v", xhttp)
	}
}

func TestVLESSLinkRoundTrip(t *testing.T) {
	link := "vless://00000000-0000-4000-8000-000000000000@example.com:443?security=reality&type=xhttp&mode=packet-up&path=%2Ftest&host=example.com&alpn=h2&encryption=mlkem768x25519plus.native.1rtt.key&support-x25519mlkem768=false#example"
	v, err := DecodeVLESSURL(link)
	if err != nil {
		t.Fatal(err)
	}
	again, err := DecodeVLESSURL(EncodeVLESSURL(v))
	if err != nil {
		t.Fatal(err)
	}
	if again.Name != "example" || again.Query.Encryption != v.Query.Encryption || again.Query.Mode != "packet-up" || again.Query.SupportX25519MLKEM768 == nil || *again.Query.SupportX25519MLKEM768 {
		t.Fatalf("lost VLESS parameters on round trip: %+v", again.Query)
	}
}

func TestUnsupportedRealitySignatureFailsClosed(t *testing.T) {
	link := "vless://00000000-0000-4000-8000-000000000000@example.com:443?security=reality&pqv=verifykey#test"
	if _, err := EncodeClash([]string{link}, SqlConfig{}); err == nil {
		t.Fatal("expected an error instead of silently dropping ML-DSA verification")
	}
}
