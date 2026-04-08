package cli

import (
	"fmt"

	"github.com/jarmocluyse/ads-go/pkg/ads"
)

// handleReadValue reads a generic value from the PLC.
// Usage: read_value [symbol_path] [port]
func handleReadValue(args []string, client *ads.Client) {
	data := "Global.int_var"
	var port uint16 = 852

	// Parse arguments if provided
	if len(args) > 0 {
		data = args[0]
	}
	if len(args) > 1 {
		fmt.Sscanf(args[1], "%d", &port)
	}

	value, err := client.ReadValue(port, data)
	if err != nil {
		fmt.Printf("[ERROR] Command 'read_value': Failed to read value from '%s' (port %d): %v\n", data, port, err)
		return
	}
	fmt.Printf("[OK] Value read from '%s' (port %d): %v\n", data, port, value)
}

// handleReadBool reads a boolean value from the PLC.
// Usage: read_bool [symbol_path] [port]
func handleReadBool(args []string, client *ads.Client) {
	data := "Global.bool_var"
	var port uint16 = 852

	// Parse arguments if provided
	if len(args) > 0 {
		data = args[0]
	}
	if len(args) > 1 {
		fmt.Sscanf(args[1], "%d", &port)
	}

	value, err := client.ReadValue(port, data)
	if err != nil {
		fmt.Printf("[ERROR] Command 'read_bool': Failed to read bool from '%s' (port %d): %v\n", data, port, err)
		return
	}
	fmt.Printf("[OK] Bool value read from '%s' (port %d): %v\n", data, port, value)
}

// handleReadObject reads a structured object from the PLC.
// Usage: read_object
func handleReadObject(args []string, client *ads.Client) {
	data := "Global.test_struct"
	var port uint16 = 852
	value, err := client.ReadValue(port, data)
	if err != nil {
		fmt.Printf("[ERROR] Command 'read_object': Failed to read object from '%s' (port %d): %v\n", data, port, err)
		return
	}
	fmt.Printf("[OK] Object value read from '%s' (port %d): %v\n", data, port, value)
}

// handleReadArray reads an array from the PLC.
// Usage: read_array
func handleReadArray(args []string, client *ads.Client) {
	data := "Global.int_array"
	var port uint16 = 852
	value, err := client.ReadValue(port, data)
	if err != nil {
		fmt.Printf("[ERROR] Command 'read_array': Failed to read array from '%s' (port %d): %v\n", data, port, err)
		return
	}
	fmt.Printf("[OK] Array value read from '%s' (port %d): %v\n", data, port, value)
}

// handleListSymbols lists all available symbols from the PLC.
// Usage: list_symbols [port]
func handleListSymbols(args []string, client *ads.Client) {
	var port uint16 = 852

	if len(args) > 0 {
		fmt.Sscanf(args[0], "%d", &port)
	}

	symbols, err := client.UploadSymbols(port)
	if err != nil {
		fmt.Printf("[ERROR] list_symbols: failed to upload symbols (port %d): %v\n", port, err)
		return
	}

	const limit = 100
	shown := len(symbols)
	if shown > limit {
		shown = limit
	}
	fmt.Printf("[OK] %d symbols total (showing first %d):\n", len(symbols), shown)
	for i, sym := range symbols[:shown] {
		fmt.Printf("  %3d: %-50s  type=%-20s  size=%d\n", i+1, sym.Name, sym.Type, sym.Size)
	}
	if len(symbols) > limit {
		fmt.Printf("  ... and %d more\n", len(symbols)-limit)
	}
}
