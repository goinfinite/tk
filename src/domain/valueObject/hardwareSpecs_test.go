package tkValueObject

import "testing"

func TestNewHardwareSpecs(t *testing.T) {
	t.Run("NewHardwareSpecs", func(t *testing.T) {
		cpuModelName, _ := NewCpuModelName("Intel Xeon E5-2670 v3")
		memoryTotal, _ := NewGibibyte(16)
		storageTotal, _ := NewGibibyte(512)

		hardwareSpecs := NewHardwareSpecs(
			cpuModelName, 8, 2.3, memoryTotal, storageTotal,
		)

		if hardwareSpecs.CpuModelName != cpuModelName {
			t.Errorf("UnexpectedCpuModelName: '%v'", hardwareSpecs.CpuModelName)
		}
		if hardwareSpecs.CpuCoresCount != 8 {
			t.Errorf("UnexpectedCpuCoresCount: '%v'", hardwareSpecs.CpuCoresCount)
		}
		if hardwareSpecs.CpuFrequencyGhz != 2.3 {
			t.Errorf("UnexpectedCpuFrequencyGhz: '%v'", hardwareSpecs.CpuFrequencyGhz)
		}
		if hardwareSpecs.MemoryTotalBytes != memoryTotal {
			t.Errorf("UnexpectedMemoryTotalBytes: '%v'", hardwareSpecs.MemoryTotalBytes)
		}
		if hardwareSpecs.StorageTotalBytes != storageTotal {
			t.Errorf("UnexpectedStorageTotalBytes: '%v'", hardwareSpecs.StorageTotalBytes)
		}
	})

	t.Run("StringMethod", func(t *testing.T) {
		memoryTotal, _ := NewGibibyte(16)

		testCaseStructs := []struct {
			inputValue     HardwareSpecs
			expectedOutput string
		}{
			{
				NewHardwareSpecs(
					CpuModelName("Intel Xeon E5-2670 v3"), 8, 2.3,
					memoryTotal, Byte(0),
				),
				"Intel Xeon E5-2670 v3 (8c@2.3 GHz) ‖ 16 GiB RAM",
			},
			{
				NewHardwareSpecs(
					CpuModelName("AMD Ryzen 9 5950X"), 16, 3.4,
					memoryTotal, Byte(0),
				),
				"AMD Ryzen 9 5950X (16c@3.4 GHz) ‖ 16 GiB RAM",
			},
			{
				NewHardwareSpecs(
					CpuModelName("Intel Xeon E5-2670 v3 2.30GHz"), 8, 2.3,
					memoryTotal, Byte(0),
				),
				"Intel Xeon E5-2670 v3 (8c@2.3 GHz) ‖ 16 GiB RAM",
			},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
