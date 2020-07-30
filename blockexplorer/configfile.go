package blockexplorer

// ConfigFile is the blockexplorer configuration
type ConfigFile struct {
	Root       string `json:"root"`
	HTTPAddr   string `json:"grpc_addr"`
	LedgerAddr string `json:"ledger_addr"`
	StatusAddr string `json:"status_addr"`
	Insecure   bool   `json:"insecure"`
}
