# Example TwinCAT Project

This directory contains a complete TwinCAT 3 project that demonstrates the capabilities of the ads-go library and CLI tool.

## Project Structure

```
example2/
├── AdsGo.sln                 # Visual Studio solution file
└── src/
    ├── AdsGo.tsproj          # TwinCAT project file
    ├── AdsGo_Test.plcproj    # TwinCAT PLC project file
    ├── Programs/
    │   └── Main.TcPOU        # Main program with cycle-based and time-based logic
    ├── Lib/
    │   └── TestStruct.TcDUT  # Sample structured data type
    └── Globals/
        └── Global.TcGVL # Global variable list with test variables
```

## Opening the Project

1. Open TwinCAT XAE (Extended Automation Engineering)
2. Open `example2/AdsGo.sln`
3. Start TwinCAT Usermode Runtime
3. Build the project: **TwinCAT → Activate Configuration**
4. Set TwinCAT to RUN mode

**Note:** The `_Boot/` directory (containing compiled boot files) is excluded from Git per TwinCAT best practices. It will be automatically generated when you activate the configuration in step 3.

## Available Variables

The project exposes the following global variables for ADS access:

### Basic Types (For Testing)
- `Global.int_var: INT` - Simple integer value (default: 0)
- `Global.bool_var: BOOL` - Simple boolean value (default: FALSE)
- `Global.dint_var: DINT` - Simple double integer value (default: 0)

### Cycle-Based Variables (Updates Every PLC Scan)
- `Global.int_counter: INT` - Increments every PLC cycle when enabled (default: 0)
- `Global.bool_toggle: BOOL` - Toggles every PLC cycle when enabled (default: FALSE)

### Time-Based Variables (Updates Every Cycle Period)
- `Global.timer_int_counter: INT` - Increments every cycle period when enabled (default: 0)
- `Global.timed_bool_toggle: BOOL` - Toggles every cycle period when enabled (default: FALSE)

### Control Flags
- `Global.int_counter_active: BOOL` - Enable/disable cycle-based counter (default: TRUE)
- `Global.bool_toggle_active: BOOL` - Enable/disable cycle-based toggle (default: TRUE)
- `Global.timed_int_counter_active: BOOL` - Enable/disable time-based counter (default: TRUE)
- `Global.timed_bool_toggle_active: BOOL` - Enable/disable time-based toggle (default: TRUE)

### Configuration
- `Global.cycle_period: TIME` - Period for time-based operations (default: T#2S = 2 seconds)

### Complex Types
- `Global.int_array: ARRAY[0..100] OF INT` - Array of 101 integers
- `Global.test_struct: TestStruct` - Structured data type with:
  - `counter: INT` - Integer counter field
  - `ready: BOOL` - Boolean ready flag
  - `int_array: ARRAY[0..50] OF INT` - Nested integer array

## PLC Program Logic

The `Main` program implements two types of behavior:

### 1. Cycle-Based Behavior (Every PLC Scan)
Executes on every PLC cycle (typically 1-10ms):

```iecst
// Increment counter if enabled
IF Global.int_counter_active THEN
    Global.int_counter := Global.int_counter + 1;
END_IF

// Toggle boolean if enabled
IF Global.bool_toggle_active THEN
    Global.bool_toggle := NOT Global.bool_toggle;
END_IF
```

**Use case:** Fast-changing values for testing rapid subscriptions and notifications.

### 2. Time-Based Behavior (Every Cycle Period)
Executes periodically based on `cycle_period` (default 2 seconds):

```iecst
// Timer triggers periodic actions
timer(IN := TRUE, PT := Global.cycle_period);

IF timer.Q THEN
    // Increment timed counter if enabled
    IF Global.timed_int_counter_active THEN
        Global.timer_int_counter := Global.timer_int_counter + 1;
    END_IF
    
    // Toggle timed boolean if enabled
    IF Global.timed_bool_toggle_active THEN
        Global.timed_bool_toggle := NOT Global.timed_bool_toggle;
    END_IF
    
    // Retrigger timer for next cycle
    timer(IN := FALSE);
END_IF
```

**Use case:** Slower-changing values for testing different subscription cycle times and demonstrating configurable periods.

## Using with ads-go CLI

The CLI tool is pre-configured to work with these variables. See the main [README.md](../README.md#example-twincat-project) for CLI commands.

### Quick Start Examples

**Monitor fast-changing counter:**
```bash
./ads-cli
> sub_counter
# Watch counter increment every PLC cycle
```

**Control the toggles:**
```bash
> enable_toggle false     # Disable cycle-based toggle
> enable_timed_toggle true  # Enable time-based toggle
> read_status             # Check current states
```

**Change timer period:**
```bash
> read_period             # Check current period (default 2s)
> set_period 5            # Change to 5 seconds
> sub_timed_counter       # Watch it increment every 5s
```

**Test all subscriptions:**
```bash
> sub_all                 # Subscribe to all 4 counters/toggles
> list_subs               # View statistics
> unsubscribe_all         # Clean up
```

## Variable Access Paths

When using ads-go library programmatically:

```go
// Read basic values
value, _ := client.ReadValue(851, "Global.int_var")
value, _ := client.ReadValue(851, "Global.bool_var")

// Read counters
counter, _ := client.ReadValue(851, "Global.int_counter")
timedCounter, _ := client.ReadValue(851, "Global.timer_int_counter")

// Write control flags
client.WriteValue(851, "Global.int_counter_active", true)
client.WriteValue(851, "Global.cycle_period", int32(5000)) // 5 seconds in ms

// Read array
array, _ := client.ReadValue(851, "Global.int_array")

// Read/write struct
dut, _ := client.ReadValue(851, "Global.test_struct")
client.WriteValue(851, "Global.test_struct", map[string]any{
    "counter": int16(100),
    "ready":   true,
})

// Subscribe to changes
callback := func(data ads.SubscriptionData) {
    fmt.Printf("Value changed: %v\n", data.Value)
}
settings := ads.SubscriptionSettings{
    CycleTime:    100 * time.Millisecond,
    SendOnChange: true,
}
sub, _ := client.SubscribeValue(851, "Global.bool_toggle", callback, settings)
```

## Requirements

- TwinCAT 3 XAE or XAR (Build 4024 or newer recommended)
- Windows OS with TwinCAT runtime
- ADS route configured between client and PLC (see main README for setup instructions)

## Port Configuration

- **Default PLC Port:** 851 (TwinCAT 3 first runtime)
- For TwinCAT 2, use port 801

## Troubleshooting

**Variables not found:**
- Ensure TwinCAT is in RUN mode (not CONFIG)
- Verify the PLC program is activated
- Check ADS route is configured correctly

**Counters not incrementing:**
- Check enable flags: `read_status` in CLI
- Verify PLC is running: `state` in CLI
- For timed counters, verify cycle period: `read_period` in CLI

**Subscription notifications not received:**
- Ensure PLC is in RUN mode
- Verify the variable is actually changing (check enable flags)
- Check subscription settings (cycle time, on-change mode)

## License

This example project is part of ads-go and is licensed under the MIT License.
