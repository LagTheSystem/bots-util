package config

type DesktopConfig struct {
	TargetFolder string
	ZipURL       string
	TempDir      string
}

type ChromeConfig struct {
	ExtensionID  string
	UpdateURL    string
	ForceInstall bool
}

type RepairConfig struct {
	Commands []RepairCommand
}

type RepairCommand struct {
	Name    string
	Exe     string
	Args    []string
	Timeout string // e.g. "30m"
}

type PolicyConfig struct {
	RegistryEntries []RegistryEntry
}

type RegistryEntry struct {
	KeyPath string
	Name    string
	Value   any
	Type    uint32 // registry.DWORD, registry.SZ, etc.
}

func DefaultDesktopConfig() DesktopConfig {
	return DesktopConfig{
		TargetFolder: `C:\Users\Public\Desktop\CampShortcuts`,
		ZipURL:       "",
		TempDir:      `C:\Windows\Temp\bots-util`,
	}
}

func DefaultChromeConfig() ChromeConfig {
	return ChromeConfig{
		ExtensionID:  "",
		UpdateURL:    "https://clients2.google.com/service/update2/crx",
		ForceInstall: true,
	}
}

func DefaultRepairConfig() RepairConfig {
	return RepairConfig{
		Commands: []RepairCommand{
			{
				Name:    "DISM RestoreHealth",
				Exe:     "dism",
				Args:    []string{"/Online", "/Cleanup-Image", "/RestoreHealth"},
				Timeout: "30m",
			},
			{
				Name:    "SFC ScanNow",
				Exe:     "sfc",
				Args:    []string{"/scannow"},
				Timeout: "15m",
			},
		},
	}
}

func DefaultPolicyConfig() PolicyConfig {
	return PolicyConfig{
		RegistryEntries: []RegistryEntry{},
	}
}
