package custom_bindings

import (
	"fmt"
	"fyne.io/fyne/v2/data/binding"
	"testing"
	"time"
)

const testTimeout = 100 * time.Millisecond

var uutFloatToStringBinding binding.String
var refFloatBinding binding.Float
var defaultFloatBinding binding.Float

func init() {
	refFloatBinding = binding.NewFloat()
	defaultFloatBinding = binding.NewFloat()
	uutFloatToStringBinding = NewLazyFloatToString(refFloatBinding, defaultFloatBinding)
}

func TestFloatToStringSetFloat(t *testing.T) {
	t.Log("WARNING: Test with inconsistent determinacy") //TODO remove if better solution found
	const testVal float64 = 10.04
	var notified = false
	var listener = binding.NewDataListener(func() {
		notified = true
	})
	uutFloatToStringBinding.AddListener(listener)

	if err := refFloatBinding.Set(testVal); err != nil {
		t.Skip("Failed to set reference.")
	}
	time.Sleep(testTimeout)
	if val, err := uutFloatToStringBinding.Get(); err != nil {
		t.Errorf("TestFloatToStringSetFloat() failed. Failed to read from Binding.")
	} else {
		if val != fmt.Sprint(testVal) {
			t.Errorf("TestFloatToStringSetFloat() failed. Failed read wrong value from bilding: Expected %s, got %s after %dms.", fmt.Sprint(testVal), val, testTimeout/time.Millisecond)
		}
	}
	if !notified {
		t.Errorf("TestFloatToStringSetFloat() failed. Listener was not notified on value change after %dms.", testTimeout/time.Millisecond)
	}
}
func TestFloatToStringSetString(t *testing.T) {
	t.Log("WARNING: Test with inconsistent determinacy") //TODO remove if better solution found
	const testVal float64 = 10.04
	var notified = false
	var listener = binding.NewDataListener(func() {
		notified = true
	})
	refFloatBinding.AddListener(listener)

	if err := uutFloatToStringBinding.Set(fmt.Sprint(10.04)); err != nil {
		t.Errorf("TestFloatToStringSetString() failed. Failed to Set Binding.")
	}
	time.Sleep(testTimeout)
	if val, err := refFloatBinding.Get(); err != nil {
		t.Errorf("TestFloatToStringSetString() failed. Failed to read from reference Binding.")
	} else {
		if val != testVal {
			t.Errorf("TestFloatToStringSetString() failed. Failed read wrong value from bilding: Expected %f, got %f after %dms.", testVal, val, testTimeout/time.Millisecond)
		}
	}
	if !notified {
		t.Errorf("TestFloatToStringSetString() failed. Reference Listener was not notified on value change after %dms.", testTimeout/time.Millisecond)
	}
}

func TestFloatToStringUseDefault(t *testing.T) {
	t.Log("WARNING: Test with inconsistent determinacy") //TODO remove if better solution found
	const defaultVal float64 = 42.911
	var notified = false
	var listener = binding.NewDataListener(func() {
		notified = true
	})
	refFloatBinding.AddListener(listener)
	if err := defaultFloatBinding.Set(defaultVal); err != nil {
		t.Skip("Failed to set default.")
	}
	if err := uutFloatToStringBinding.Set("EinHoffentlichNichtParsbarerSatz!"); err == nil {
		t.Errorf("TestFloatToStringUseDefault() failed. Binding Set accepted input it should not.")
	}
	time.Sleep(testTimeout)
	if !notified {
		t.Errorf("TestFloatToStringUseDefault() failed. Failed to notify reference listener.")
	}
	if val, err := refFloatBinding.Get(); err != nil {
		t.Errorf("TestFloatToStringUseDefault() failed. Failed to read from reference Binding.")
	} else {
		if val != defaultVal {
			t.Errorf("TestFloatToStringUseDefault() failed. Failed read wrong value from bilding: Expected %f, got %f after %dms.", defaultVal, val, testTimeout/time.Millisecond)
		}
	}

	if err := uutFloatToStringBinding.Set(""); err != nil {
		t.Errorf("TestFloatToStringUseDefault() failed. Failed to Set Binding.")
	}
	notified = false
	time.Sleep(testTimeout)
	if notified {
		t.Errorf("TestFloatToStringUseDefault() failed. Notified reference listener when it should not.")
	}
	if val, err := refFloatBinding.Get(); err != nil {
		t.Errorf("TestFloatToStringUseDefault() failed. Failed to read from reference Binding.")
	} else {
		if val != defaultVal {
			t.Errorf("TestFloatToStringUseDefault() failed. Failed read wrong value from bilding: Expected %f, got %f after %dms.", defaultVal, val, testTimeout/time.Millisecond)
		}
	}

}
