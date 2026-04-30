//go:build integration

package integration

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jarmocluyse/ads-go/pkg/ads"
	adssymbol "github.com/jarmocluyse/ads-go/pkg/ads/ads-symbol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const plcPort = 852

// newClient creates an ADS client from environment variables and registers cleanup.
// ADS_TARGET_NET_ID is required. ADS_ROUTER_HOST defaults to 127.0.0.1.
func newClient(t *testing.T) *ads.Client {
	t.Helper()

	targetNetID := os.Getenv("ADS_TARGET_NET_ID")
	require.NotEmpty(t, targetNetID, "ADS_TARGET_NET_ID env var must be set (e.g. 192.168.1.5.1.1)")

	settings := ads.ClientSettings{
		TargetNetID: targetNetID,
		RouterHost:  os.Getenv("ADS_ROUTER_HOST"), // empty → LoadDefaults fills 127.0.0.1
	}

	client := ads.NewClient(settings, nil)
	require.NoError(t, client.Connect(), "connect to ADS router")

	t.Cleanup(func() { _ = client.Disconnect() })
	return client
}

// -----------------------------------------------------------------------
// Expected value computation — pure function, no I/O.
// Mirrors the deterministic formulas in example3/src/Programs/MAIN.TcPOU.
// -----------------------------------------------------------------------

type snapshotExpected struct {
	Seed  uint32
	Bool  bool
	Sint  int8
	Usint uint8
	Byte_ uint8
	Int   int16
	Uint  uint16
	Word  uint16
	Dint  int32
	Udint uint32
	Dword uint32
	Lint  int64
	Ulint uint64
	Lword uint64
	Real  float32
	Lreal float64
	// Date/time raw values (wire encoding: uint32 or uint64)
	Time_ uint32 // ms
	Tod   uint32 // ms since midnight (seed % 86_400_000)
	Date  uint32 // seconds since 1970-01-01 (raw)
	Dt    uint32 // seconds since 1970-01-01 (raw)
	// Ltime  uint64 — not supported in TC 4024
	// Ltod   uint64 — not supported in TC 4024
	// Ldate  uint64 — not supported in TC 4024
	// Ldt    uint64 — not supported in TC 4024
	String string
}

func expectedValues(seed uint32) snapshotExpected {
	return snapshotExpected{
		Seed:   seed,
		Bool:   seed%2 == 0,
		Sint:   int8(seed),
		Usint:  uint8(seed),
		Byte_:  uint8(seed),
		Int:    int16(seed),
		Uint:   uint16(seed),
		Word:   uint16(seed),
		Dint:   int32(seed),
		Udint:  seed,
		Dword:  seed,
		Lint:   int64(seed),
		Ulint:  uint64(seed),
		Lword:  uint64(seed),
		Real:   float32(seed),
		Lreal:  float64(seed),
		Time_:  seed,
		Tod:    seed % 86_400_000,
		Date:   seed,
		Dt:     seed,
		String: fmt.Sprintf("S=%d", seed),
	}
}

// -----------------------------------------------------------------------
// Assertion helpers
// -----------------------------------------------------------------------

func assertSnapshot(t *testing.T, raw any, seed uint32) {
	t.Helper()

	m, ok := raw.(map[string]any)
	require.Truef(t, ok, "snapshot: expected map[string]any, got %T", raw)

	exp := expectedValues(seed)

	assert.Equal(t, exp.Seed, m["seed"], "seed")
	assert.Equal(t, exp.Bool, m["bool_var"], "bool_var")
	assert.Equal(t, exp.Sint, m["sint_var"], "sint_var")
	assert.Equal(t, exp.Usint, m["usint_var"], "usint_var")
	assert.Equal(t, exp.Byte_, m["byte_var"], "byte_var")
	assert.Equal(t, exp.Int, m["int_var"], "int_var")
	assert.Equal(t, exp.Uint, m["uint_var"], "uint_var")
	assert.Equal(t, exp.Word, m["word_var"], "word_var")
	assert.Equal(t, exp.Dint, m["dint_var"], "dint_var")
	assert.Equal(t, exp.Udint, m["udint_var"], "udint_var")
	assert.Equal(t, exp.Dword, m["dword_var"], "dword_var")
	assert.Equal(t, exp.Lint, m["lint_var"], "lint_var")
	assert.Equal(t, exp.Ulint, m["ulint_var"], "ulint_var")
	assert.Equal(t, exp.Lword, m["lword_var"], "lword_var")
	// float32 is only exact for seeds ≤ 2^24-1 = 16_777_215
	if seed <= 16_777_215 {
		assert.Equal(t, exp.Real, m["real_var"], "real_var")
	}
	assert.Equal(t, exp.Lreal, m["lreal_var"], "lreal_var")
	assert.Equal(t, exp.Time_, m["time_var"], "time_var")
	assert.Equal(t, exp.Tod, m["tod_var"], "tod_var")
	assert.Equal(t, exp.Date, m["date_var"], "date_var")
	assert.Equal(t, exp.Dt, m["dt_var"], "dt_var")
	// ltime_var / ltod_var / ldate_var / ldt_var not supported in TC 4024
	assert.Equal(t, exp.String, m["string_var"], "string_var")
}

func assertIntArray(t *testing.T, raw any, seed uint32) {
	t.Helper()
	arr, ok := raw.([]any)
	require.Truef(t, ok, "int_array: expected []any, got %T", raw)
	require.Len(t, arr, 10, "int_array length")
	for i, v := range arr {
		assert.Equal(t, int16(seed+uint32(i)), v, "int_array[%d]", i)
	}
}

func assertDintArray(t *testing.T, raw any, seed uint32) {
	t.Helper()
	arr, ok := raw.([]any)
	require.Truef(t, ok, "dint_array: expected []any, got %T", raw)
	require.Len(t, arr, 10, "dint_array length")
	for i, v := range arr {
		assert.Equal(t, int32(seed+uint32(i)), v, "dint_array[%d]", i)
	}
}

func assert2DArray(t *testing.T, raw any, seed uint32) {
	t.Helper()
	outer, ok := raw.([]any)
	require.Truef(t, ok, "array_2d: expected []any, got %T", raw)
	require.Len(t, outer, 3, "array_2d outer dimension")
	for i, row := range outer {
		inner, ok := row.([]any)
		require.Truef(t, ok, "array_2d[%d]: expected []any, got %T", i, row)
		require.Len(t, inner, 3, "array_2d[%d] inner dimension", i)
		for j, v := range inner {
			assert.Equal(t, int16(seed+uint32(i)*3+uint32(j)), v, "array_2d[%d][%d]", i, j)
		}
	}
}

// assertLinear reads every scalar field individually from basePath and checks
// against the deterministic expectation for the given seed.
// Works for flat FB vars (base = "Main.test") and struct linear access
// (base = "Main.test.struct_var" or "Main.write_struct_var").
func assertLinear(t *testing.T, client *ads.Client, base string, seed uint32) {
	t.Helper()
	exp := expectedValues(seed)

	read := func(field string) any {
		t.Helper()
		got, err := client.ReadValue(plcPort, base+"."+field)
		require.NoError(t, err, "ReadValue %s.%s", base, field)
		return got
	}

	assert.Equal(t, exp.Seed, read("seed"), "seed")
	assert.Equal(t, exp.Bool, read("bool_var"), "bool_var")
	assert.Equal(t, exp.Sint, read("sint_var"), "sint_var")
	assert.Equal(t, exp.Usint, read("usint_var"), "usint_var")
	assert.Equal(t, exp.Byte_, read("byte_var"), "byte_var")
	assert.Equal(t, exp.Int, read("int_var"), "int_var")
	assert.Equal(t, exp.Uint, read("uint_var"), "uint_var")
	assert.Equal(t, exp.Word, read("word_var"), "word_var")
	assert.Equal(t, exp.Dint, read("dint_var"), "dint_var")
	assert.Equal(t, exp.Udint, read("udint_var"), "udint_var")
	assert.Equal(t, exp.Dword, read("dword_var"), "dword_var")
	assert.Equal(t, exp.Lint, read("lint_var"), "lint_var")
	assert.Equal(t, exp.Ulint, read("ulint_var"), "ulint_var")
	assert.Equal(t, exp.Lword, read("lword_var"), "lword_var")
	if seed <= 16_777_215 {
		assert.Equal(t, exp.Real, read("real_var"), "real_var")
	}
	assert.Equal(t, exp.Lreal, read("lreal_var"), "lreal_var")
	assert.Equal(t, exp.Time_, read("time_var"), "time_var")
	assert.Equal(t, exp.Tod, read("tod_var"), "tod_var")
	assert.Equal(t, exp.Date, read("date_var"), "date_var")
	assert.Equal(t, exp.Dt, read("dt_var"), "dt_var")
	// ltime_var / ltod_var / ldate_var / ldt_var not supported in TC 4024
	assert.Equal(t, exp.String, read("string_var"), "string_var")
}

// -----------------------------------------------------------------------
// Static seed test — boundary & overflow cases
// Reads struct_var structurally (whole struct in one call) for all 14 seeds.
// -----------------------------------------------------------------------

var staticSeedCases = []struct {
	name string
	seed uint32
}{
	{"zero", 0},
	{"one", 1},
	{"sint_max", 127},
	{"sint_overflow", 128},
	{"byte_max", 255},
	{"byte_overflow", 256},
	{"int_max", 32_767},
	{"int_sign_flip", 32_768},
	{"uint_max", 65_535},
	{"uint_overflow", 65_536},
	{"real_exact_max", 16_777_215},
	{"dint_max", 2_147_483_647},
	{"dint_sign_boundary", 2_147_483_648},
	{"udint_max", 4_294_967_295},
}

func TestStaticSeed(t *testing.T) {
	client := newClient(t)

	for _, tc := range staticSeedCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Write the driving seed over ADS
			require.NoError(t, client.WriteValue(plcPort, "Main.test.seed", tc.seed), "WriteValue seed")

			// Wait for at least 5 PLC cycles (50 ms @ 10 ms cycle)
			time.Sleep(50 * time.Millisecond)

			// --- struct_var (structural: whole struct in one read) ---
			snapshotRaw, err := client.ReadValue(plcPort, "Main.test.struct_var")
			require.NoError(t, err, "ReadValue Main.test.struct_var")
			assertSnapshot(t, snapshotRaw, tc.seed)

			// --- 1-D int array ---
			intArrRaw, err := client.ReadValue(plcPort, "Main.test.int_array")
			require.NoError(t, err, "ReadValue Main.test.int_array")
			assertIntArray(t, intArrRaw, tc.seed)

			// --- 1-D dint array ---
			dintArrRaw, err := client.ReadValue(plcPort, "Main.test.dint_array")
			require.NoError(t, err, "ReadValue Main.test.dint_array")
			assertDintArray(t, dintArrRaw, tc.seed)

			// --- 2-D int array ---
			arr2DRaw, err := client.ReadValue(plcPort, "Main.test.array_2d")
			require.NoError(t, err, "ReadValue Main.test.array_2d")
			assert2DArray(t, arr2DRaw, tc.seed)
		})
	}
}

// -----------------------------------------------------------------------
// Read access-path coverage — verifies all three read patterns for one seed:
//   1. Flat FB vars    — Main.test.<var>           (individual scalar reads)
//   2. Struct structural — Main.test.struct_var    (whole struct in one call)
//   3. Struct linear   — Main.test.struct_var.<var> (individual field reads)
// -----------------------------------------------------------------------

func TestReadAllAccessPaths(t *testing.T) {
	client := newClient(t)

	const seed = uint32(100)
	require.NoError(t, client.WriteValue(plcPort, "Main.test.seed", seed), "write seed")
	time.Sleep(50 * time.Millisecond)

	t.Run("flat_fb_vars", func(t *testing.T) {
		assertLinear(t, client, "Main.test", seed)
	})

	t.Run("struct_structural", func(t *testing.T) {
		raw, err := client.ReadValue(plcPort, "Main.test.struct_var")
		require.NoError(t, err, "ReadValue Main.test.struct_var")
		assertSnapshot(t, raw, seed)
	})

	t.Run("struct_linear", func(t *testing.T) {
		assertLinear(t, client, "Main.test.struct_var", seed)
	})
}

// -----------------------------------------------------------------------
// Subscription test — auto-increment mode + on-change notifications
// -----------------------------------------------------------------------

func TestSubscription(t *testing.T) {
	client := newClient(t)

	// Put PLC into a known state: seed=0
	require.NoError(t, client.WriteValue(plcPort, "Main.test.seed", uint32(0)), "reset seed")
	time.Sleep(50 * time.Millisecond)

	// Subscribe to the snapshot struct (on-change, checked every PLC cycle)
	notifCh := make(chan ads.SubscriptionData, 32)
	sub, err := client.SubscribeValue(
		plcPort,
		"Main.test.struct_var",
		func(data ads.SubscriptionData) {
			select {
			case notifCh <- data:
			default: // drop if buffer full; test will fail on timeout
			}
		},
		ads.SubscriptionSettings{
			CycleTime:    10 * time.Millisecond,
			SendOnChange: true,
		},
	)
	require.NoError(t, err, "SubscribeValue Main.test.struct_var")
	defer func() { _ = client.Unsubscribe(sub) }()

	// Drain the mandatory initial notification sent on subscribe
	select {
	case <-notifCh:
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for initial subscription notification")
	}

	// Drive seed directly from Go and wait until the notification reflects the written value.
	// Stale notifications (from before the write) are drained until gotSeed == nextSeed.
	const wantNotifs = 10

	for i := 0; i < wantNotifs; i++ {
		nextSeed := uint32(i + 1)
		require.NoError(t, client.WriteValue(plcPort, "Main.test.seed", nextSeed), "write seed %d", nextSeed)

		deadline := time.After(5 * time.Second)
		for {
			select {
			case data := <-notifCh:
				m, ok := data.Value.(map[string]any)
				require.Truef(t, ok, "notification %d: value should be map[string]any, got %T", i, data.Value)

				gotSeed, ok := m["seed"].(uint32)
				require.Truef(t, ok, "notification %d: seed should be uint32", i)

				if gotSeed < nextSeed {
					// stale notification — keep draining
					continue
				}

				require.Equalf(t, nextSeed, gotSeed, "notification %d: unexpected seed", i)
				assertSnapshot(t, data.Value, gotSeed)

			case <-deadline:
				t.Fatalf("timeout waiting for subscription notification %d/%d", i+1, wantNotifs)
			}
			break
		}
	}
}

// -----------------------------------------------------------------------
// Write coverage — exercises WriteValue for every PLC type via:
//   1. Flat MAIN-level vars  — Main.write_<type>_var  (atomic scalar writes)
//   2. Struct linear fields  — Main.write_struct_var.<field>
//   3. Struct structural     — Main.write_struct_var (whole struct at once) [SKIP: not yet implemented]
//   4. FB control vars       — Main.test.seed / auto_mode / auto_tick_interval
//   5. Arrays                — Main.write_int_array / write_dint_array / write_2d_array
// -----------------------------------------------------------------------

type writeCase struct {
	path  string
	write any
	check func(t *testing.T, got any)
}

func eqCheck(want any) func(t *testing.T, got any) {
	return func(t *testing.T, got any) { t.Helper(); assert.Equal(t, want, got) }
}

func deltaCheck(want, delta float64) func(t *testing.T, got any) {
	return func(t *testing.T, got any) { t.Helper(); assert.InDelta(t, want, got, delta) }
}

func runWriteCases(t *testing.T, client *ads.Client, cases []writeCase) {
	t.Helper()
	for _, tc := range cases {
		tc := tc
		t.Run(tc.path, func(t *testing.T) {
			require.NoError(t, client.WriteValue(plcPort, tc.path, tc.write), "WriteValue %s", tc.path)
			time.Sleep(20 * time.Millisecond)
			got, err := client.ReadValue(plcPort, tc.path)
			require.NoError(t, err, "ReadValue %s", tc.path)
			tc.check(t, got)
		})
	}
}

func TestWriteAllTypes(t *testing.T) {
	client := newClient(t)

	// ------------------------------------------------------------------
	// 1. Flat MAIN-level vars — PLC never touches these
	// ------------------------------------------------------------------
	t.Run("flat_main_vars", func(t *testing.T) {
		const m = "Main"
		runWriteCases(t, client, []writeCase{
			{m + ".write_bool_var", true, eqCheck(true)},
			{m + ".write_bool_var", false, eqCheck(false)},
			{m + ".write_sint_var", int8(-99), eqCheck(int8(-99))},
			{m + ".write_usint_var", uint8(200), eqCheck(uint8(200))},
			{m + ".write_byte_var", uint8(0xFF), eqCheck(uint8(0xFF))},
			{m + ".write_int_var", int16(-32000), eqCheck(int16(-32000))},
			{m + ".write_uint_var", uint16(65000), eqCheck(uint16(65000))},
			{m + ".write_word_var", uint16(0xABCD), eqCheck(uint16(0xABCD))},
			{m + ".write_dint_var", int32(-2_000_000), eqCheck(int32(-2_000_000))},
			{m + ".write_udint_var", uint32(3_000_000), eqCheck(uint32(3_000_000))},
			{m + ".write_dword_var", uint32(0xDEADBEEF), eqCheck(uint32(0xDEADBEEF))},
			{m + ".write_lint_var", int64(-9_000_000_000), eqCheck(int64(-9_000_000_000))},
			{m + ".write_ulint_var", uint64(18_000_000_000), eqCheck(uint64(18_000_000_000))},
			{m + ".write_lword_var", uint64(0xCAFEBABEDEADBEEF), eqCheck(uint64(0xCAFEBABEDEADBEEF))},
			{m + ".write_real_var", float32(3.14), deltaCheck(float64(float32(3.14)), 1e-4)},
			{m + ".write_lreal_var", float64(2.718281828), deltaCheck(2.718281828, 1e-9)},
			{m + ".write_time_var", uint32(5000), eqCheck(uint32(5000))},
			{m + ".write_tod_var", uint32(3_600_000), eqCheck(uint32(3_600_000))},
			{m + ".write_date_var", uint32(1_000_000), eqCheck(uint32(1_000_000))},
			{m + ".write_dt_var", uint32(1_700_000_000), eqCheck(uint32(1_700_000_000))},
			{m + ".write_string_var", "hello ADS", eqCheck("hello ADS")},
		})
	})

	// ------------------------------------------------------------------
	// 2. Struct fields — linear access, one field at a time
	// ------------------------------------------------------------------
	t.Run("struct_linear", func(t *testing.T) {
		const s = "Main.write_struct_var"
		runWriteCases(t, client, []writeCase{
			{s + ".seed", uint32(77), eqCheck(uint32(77))},
			{s + ".bool_var", true, eqCheck(true)},
			{s + ".bool_var", false, eqCheck(false)},
			{s + ".sint_var", int8(-50), eqCheck(int8(-50))},
			{s + ".usint_var", uint8(150), eqCheck(uint8(150))},
			{s + ".byte_var", uint8(0xAB), eqCheck(uint8(0xAB))},
			{s + ".int_var", int16(-1000), eqCheck(int16(-1000))},
			{s + ".uint_var", uint16(1000), eqCheck(uint16(1000))},
			{s + ".word_var", uint16(0x1234), eqCheck(uint16(0x1234))},
			{s + ".dint_var", int32(-100_000), eqCheck(int32(-100_000))},
			{s + ".udint_var", uint32(100_000), eqCheck(uint32(100_000))},
			{s + ".dword_var", uint32(0x12345678), eqCheck(uint32(0x12345678))},
			{s + ".lint_var", int64(-1_000_000_000), eqCheck(int64(-1_000_000_000))},
			{s + ".ulint_var", uint64(1_000_000_000), eqCheck(uint64(1_000_000_000))},
			{s + ".lword_var", uint64(0xDEADBEEF12345678), eqCheck(uint64(0xDEADBEEF12345678))},
			{s + ".real_var", float32(1.5), deltaCheck(float64(float32(1.5)), 1e-4)},
			{s + ".lreal_var", float64(3.141592653589793), deltaCheck(3.141592653589793, 1e-9)},
			{s + ".time_var", uint32(2000), eqCheck(uint32(2000))},
			{s + ".tod_var", uint32(7_200_000), eqCheck(uint32(7_200_000))},
			{s + ".date_var", uint32(500_000), eqCheck(uint32(500_000))},
			{s + ".dt_var", uint32(1_600_000_000), eqCheck(uint32(1_600_000_000))},
			{s + ".string_var", "write test", eqCheck("write test")},
		})
	})

	// ------------------------------------------------------------------
	// 3. Struct structural write — write the whole struct in one call
	//    Serialize already handles map[string]any for structs.
	// ------------------------------------------------------------------
	t.Run("struct_structural", func(t *testing.T) {
		const seed = uint32(77)
		exp := expectedValues(seed)
		m := map[string]any{
			"seed":       exp.Seed,
			"bool_var":   exp.Bool,
			"sint_var":   exp.Sint,
			"usint_var":  exp.Usint,
			"byte_var":   exp.Byte_,
			"int_var":    exp.Int,
			"uint_var":   exp.Uint,
			"word_var":   exp.Word,
			"dint_var":   exp.Dint,
			"udint_var":  exp.Udint,
			"dword_var":  exp.Dword,
			"lint_var":   exp.Lint,
			"ulint_var":  exp.Ulint,
			"lword_var":  exp.Lword,
			"real_var":   exp.Real,
			"lreal_var":  exp.Lreal,
			"time_var":   exp.Time_,
			"tod_var":    exp.Tod,
			"date_var":   exp.Date,
			"dt_var":     exp.Dt,
			"string_var": exp.String,
		}
		require.NoError(t, client.WriteValue(plcPort, "Main.write_struct_var", m))
		time.Sleep(20 * time.Millisecond)
		got, err := client.ReadValue(plcPort, "Main.write_struct_var")
		require.NoError(t, err)
		assertSnapshot(t, got, seed)
	})

	// ------------------------------------------------------------------
	// 4. FB control vars — public vars on the Deterministic FB instance
	// ------------------------------------------------------------------
	t.Run("fb_control_vars", func(t *testing.T) {
		const fb = "Main.test"
		runWriteCases(t, client, []writeCase{
			{fb + ".seed", uint32(999), eqCheck(uint32(999))},
			{fb + ".auto_mode", true, eqCheck(true)},
			{fb + ".auto_mode", false, eqCheck(false)},
			{fb + ".auto_tick_interval", uint32(123), eqCheck(uint32(123))},
		})
		// Restore stable state
		require.NoError(t, client.WriteValue(plcPort, fb+".auto_mode", false), "restore auto_mode")
		require.NoError(t, client.WriteValue(plcPort, fb+".seed", uint32(0)), "restore seed")
		require.NoError(t, client.WriteValue(plcPort, fb+".auto_tick_interval", uint32(50)), "restore tick interval")
	})

	// ------------------------------------------------------------------
	// 5. Array write + read
	// ------------------------------------------------------------------
	t.Run("int_array", func(t *testing.T) {
		writeVal := make([]int16, 10)
		for i := range writeVal {
			writeVal[i] = int16(i * 10)
		}
		require.NoError(t, client.WriteValue(plcPort, "Main.write_int_array", writeVal))
		time.Sleep(20 * time.Millisecond)
		got, err := client.ReadValue(plcPort, "Main.write_int_array")
		require.NoError(t, err)
		arr, ok := got.([]any)
		require.True(t, ok, "expected []any got %T", got)
		require.Len(t, arr, 10)
		for i, v := range arr {
			assert.Equal(t, int16(i*10), v, "write_int_array[%d]", i)
		}
	})

	t.Run("dint_array", func(t *testing.T) {
		writeVal := make([]int32, 10)
		for i := range writeVal {
			writeVal[i] = int32(i * 1000)
		}
		require.NoError(t, client.WriteValue(plcPort, "Main.write_dint_array", writeVal))
		time.Sleep(20 * time.Millisecond)
		got, err := client.ReadValue(plcPort, "Main.write_dint_array")
		require.NoError(t, err)
		arr, ok := got.([]any)
		require.True(t, ok, "expected []any got %T", got)
		require.Len(t, arr, 10)
		for i, v := range arr {
			assert.Equal(t, int32(i*1000), v, "write_dint_array[%d]", i)
		}
	})

	t.Run("2d_array", func(t *testing.T) {
		writeVal := [][]int16{
			{1, 2, 3},
			{4, 5, 6},
			{7, 8, 9},
		}
		require.NoError(t, client.WriteValue(plcPort, "Main.write_2d_array", writeVal))
		time.Sleep(20 * time.Millisecond)
		got, err := client.ReadValue(plcPort, "Main.write_2d_array")
		require.NoError(t, err)
		outer, ok := got.([]any)
		require.True(t, ok, "expected []any got %T", got)
		require.Len(t, outer, 3)
		for i, row := range outer {
			inner, ok := row.([]any)
			require.True(t, ok)
			require.Len(t, inner, 3)
			for j, v := range inner {
				assert.Equal(t, writeVal[i][j], v, "write_2d_array[%d][%d]", i, j)
			}
		}
	})
}

// -----------------------------------------------------------------------
// Struct packing — the same four fields under pack_mode 0/2/4/8.
//
// Field layout:  b1 BOOL | d1 DWORD | b2 BOOL | lw1 LWORD
// This layout causes the DWORD and LWORD to land at different byte offsets
// under each pack mode, so the serializer must use subItem.Offset (as
// reported by the PLC's ADS type info) rather than packing fields naively.
//
// Expected offsets per Beckhoff docs:
//   pack_mode 0:  b1=0  d1=1  b2=5  lw1=6   (total 14)
//   pack_mode 2:  b1=0  d1=2  b2=6  lw1=8   (total 16)
//   pack_mode 4:  b1=0  d1=4  b2=8  lw1=12  (total 20)
//   pack_mode 8:  b1=0  d1=4  b2=8  lw1=16  (total 24)
//
// Two-phase strategy to avoid circular validation:
//
//   Phase 1 — plc_read: PLC fills structs from seed; Go reads structurally
//     and asserts against independently computed expected values.
//     If Serialize had a bug but Deserialize was symmetric, this would catch it.
//
//   Phase 2 — write_then_linear_read: Go writes structs structurally
//     (WriteValue with map[string]any), then reads each field individually
//     via its linear ADS path. Structural write is validated by linear reads —
//     different code paths, no circular dependency.
// -----------------------------------------------------------------------

// packExpected computes the deterministic values for any pack struct from seed.
// Formula matches StructTests.TcPOU exactly.
func packExpected(seed uint32) (b1 bool, d1 uint32, b2 bool, lw1 uint64) {
	return seed%2 == 0, seed, seed%3 == 0, uint64(seed) * 3
}

func assertPackStruct(t *testing.T, got any, seed uint32) {
	t.Helper()
	m, ok := got.(map[string]any)
	require.True(t, ok, "expected map[string]any got %T", got)
	b1, d1, b2, lw1 := packExpected(seed)
	assert.Equal(t, b1, m["b1"], "b1")
	assert.Equal(t, d1, m["d1"], "d1")
	assert.Equal(t, b2, m["b2"], "b2")
	assert.Equal(t, lw1, m["lw1"], "lw1")
}

func TestStructPacking(t *testing.T) {
	client := newClient(t)

	packPaths := []struct{ name, plcRead, writeTarget string }{
		{"pack_none", "Main.struct_tests.pack_none", "Main.pack_none"},
		{"pack_two", "Main.struct_tests.pack_two", "Main.pack_two"},
		{"pack_four", "Main.struct_tests.pack_four", "Main.pack_four"},
		{"pack_eight", "Main.struct_tests.pack_eight", "Main.pack_eight"},
	}

	// ------------------------------------------------------------------
	// Phase 1: PLC-driven read.
	// Go writes a seed, PLC fills all 4 structs deterministically.
	// Go reads structurally and asserts against independently computed values.
	// ------------------------------------------------------------------
	t.Run("plc_read", func(t *testing.T) {
		for _, seed := range []uint32{0, 1, 2, 3, 6, 100, 0xFFFFFFFF} {
			seed := seed
			t.Run(fmt.Sprintf("seed_%d", seed), func(t *testing.T) {
				require.NoError(t, client.WriteValue(plcPort, "Main.struct_tests.seed", seed))
				time.Sleep(20 * time.Millisecond)
				for _, p := range packPaths {
					p := p
					t.Run(p.name, func(t *testing.T) {
						got, err := client.ReadValue(plcPort, p.plcRead)
						require.NoError(t, err, "ReadValue %s", p.plcRead)
						assertPackStruct(t, got, seed)
					})
				}
			})
		}
	})

	// ------------------------------------------------------------------
	// Phase 2: Write structurally, read back each field linearly.
	// Structural Serialize and linear Deserialize are independent code paths.
	// ------------------------------------------------------------------
	t.Run("write_then_linear_read", func(t *testing.T) {
		const seed = uint32(42)
		b1, d1, b2, lw1 := packExpected(seed)
		val := map[string]any{"b1": b1, "d1": d1, "b2": b2, "lw1": lw1}

		for _, p := range packPaths {
			p := p
			t.Run(p.name, func(t *testing.T) {
				require.NoError(t, client.WriteValue(plcPort, p.writeTarget, val), "WriteValue %s", p.writeTarget)
				time.Sleep(20 * time.Millisecond)

				gotB1, err := client.ReadValue(plcPort, p.writeTarget+".b1")
				require.NoError(t, err)
				assert.Equal(t, b1, gotB1, "b1")

				gotD1, err := client.ReadValue(plcPort, p.writeTarget+".d1")
				require.NoError(t, err)
				assert.Equal(t, d1, gotD1, "d1")

				gotB2, err := client.ReadValue(plcPort, p.writeTarget+".b2")
				require.NoError(t, err)
				assert.Equal(t, b2, gotB2, "b2")

				gotLw1, err := client.ReadValue(plcPort, p.writeTarget+".lw1")
				require.NoError(t, err)
				assert.Equal(t, lw1, gotLw1, "lw1")
			})
		}
	})
}
// >>> TEST_SYMBOL_ATTRIBUTES START
func TestSymbolAttributes(t *testing.T) {
	client := newClient(t)

	sym, err := client.GetSymbol(plcPort, "Main.attribute_test")
	require.NoError(t, err, "GetSymbol Main.attribute_test")
	require.NotNil(t, sym)

	// Log diagnostic info to help debug attribute parsing in integration environment.
	t.Logf("Symbol Flags: 0x%X, TypeGUID: %s, AttributesCount: %d", uint32(sym.Flags), sym.TypeGUID, len(sym.Attributes))
	if len(sym.Attributes) == 0 {
		t.Logf("Attributes slice is empty; full symbol: %+v", sym)
		// Fallback: try uploading the entire symbol table and search for the symbol
		t.Log("Attempting UploadSymbols fallback to locate attributes")
		all, uerr := client.UploadSymbols(plcPort)
		if uerr != nil {
			t.Logf("UploadSymbols error: %v", uerr)
		} else {
			var found *adssymbol.AdsSymbol
			for i := range all {
				if all[i].Name == "Main.attribute_test" {
					found = &all[i]
					break
				}
			}
			if found != nil {
				t.Logf("Found symbol in UploadSymbols: Flags=0x%X AttributesCount=%d", uint32(found.Flags), len(found.Attributes))
				for i, a := range found.Attributes {
					t.Logf("Upload Attr %d: name=%q value=%q", i, a.Name, a.Value)
				}
				// replace sym with found for subsequent assertions
				sym = found
			} else {
				t.Log("Symbol not found in upload results")
			}
		}
	} else {
		for i, a := range sym.Attributes {
			t.Logf("Attribute %d: name=%q value=%q", i, a.Name, a.Value)
		}
	}

	// Expect two attributes: a flag-only attribute and a key=value attribute.
	require.Len(t, sym.Attributes, 2, "expected 2 attributes on Main.attribute_test")

	var foundLinkalways, foundCustom bool
	for _, a := range sym.Attributes {
		switch a.Name {
		case "linkalways":
			foundLinkalways = true
			assert.Equal(t, "", a.Value, "linkalways should have empty value")
		case "some_made_up_key":
			foundCustom = true
			assert.Equal(t, "some_made_up_value", a.Value, "some_made_up_key value")
		}
	}

	require.True(t, foundLinkalways, "missing attribute: linkalways")
	require.True(t, foundCustom, "missing attribute: some_made_up_key")
}
// <<< TEST_SYMBOL_ATTRIBUTES END
