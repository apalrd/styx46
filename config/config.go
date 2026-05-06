package config

import (
    "os"
    "fmt"
    "net"
    "gopkg.in/yaml.v3"
)

type DownstreamIf struct {
    Name string `yaml:"name"`
    Map4 string `yaml:"map4"`
    Map6 string `yaml:"map6"`
}

type UpstreamIf struct {
    Name    string `yaml:"name"`
    ProxyND bool   `yaml:"proxy_nd"`
    TaygaIP string `yaml:"tayga_ip"`
}

type Config struct {
    Listen    	string   `yaml:"listen"`     // clients connect to us e.g. "0.0.0.0:53"
    Upstreams 	[]string `yaml:"upstreams"`  // we connect upstream e.g. ["[2620:fe::fe]:53", "9.9.9.9:53"]
	Legacy		string	 `yaml:"legacy"`	 // legacy network space reserved for this, i.e. "10.0.0.0/8"
	BinaryPath	string	 `yaml:"binary_path"` // Path or binary file name of the translator (default "tayga")
	PreferSynth	bool	 `yaml:"prefer_synth"` // Prefer to synthesize addresses if AAAA records are available
	Pref64		string	 `yaml:"pref64"` 	 // Pref64 for the PLAT
    Debug       bool     `yaml:"debug"`      // Debug-level query logging
    Pool        *net.IPNet                   // decoded from Legacy
    IfUp        UpstreamIf  `yaml:"iface_upstream"`    // interface name for upstream traffic (if defined, will do proxy-ndp)
    IfDowns     []DownstreamIf `yaml:"iface_downstream"` // downstream interfaces
}

func Load(path string) (*Config, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    var cfg Config
    if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
        return nil, err
    }

    //validate ip net
    ip, network, err := net.ParseCIDR(cfg.Legacy)
    if err != nil {
        return nil, fmt.Errorf("invalid CIDR: %w", err)
    }
    if ip.Equal(network.IP) == false {
        return nil, fmt.Errorf("CIDR must be a network address, not a host address")
    }
    cfg.Pool = network
    if cfg.BinaryPath == "" {
        cfg.BinaryPath = "tayga"
    }
    return &cfg, nil
}