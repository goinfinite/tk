package tkValueObject

import (
	"fmt"
	"strings"
)

type HardwareSpecs struct {
	CpuModelName      CpuModelName `json:"cpuModelName"`
	CpuCoresCount     float64      `json:"cpuCoresCount"`
	CpuFrequencyGhz   float64      `json:"cpuFrequencyGhz"`
	MemoryTotalBytes  Byte         `json:"memoryTotalBytes"`
	StorageTotalBytes Byte         `json:"storageTotalBytes"`
}

func NewHardwareSpecs(
	cpuModelName CpuModelName,
	cpuCoresCount, cpuFrequencyGhz float64,
	memoryTotalBytes, storageTotalBytes Byte,
) HardwareSpecs {
	return HardwareSpecs{
		CpuModelName:      cpuModelName,
		CpuCoresCount:     cpuCoresCount,
		CpuFrequencyGhz:   cpuFrequencyGhz,
		MemoryTotalBytes:  memoryTotalBytes,
		StorageTotalBytes: storageTotalBytes,
	}
}

func (vo HardwareSpecs) String() string {
	cpuModelNameParts := strings.Split(vo.CpuModelName.String(), " ")
	if len(cpuModelNameParts) > 4 {
		cpuModelNameParts = cpuModelNameParts[:4]
	}
	cpuModelNameStr := strings.Join(cpuModelNameParts, " ")

	return fmt.Sprintf(
		"%s (%.0fc@%.1f GHz) ‖ %s RAM",
		cpuModelNameStr, vo.CpuCoresCount,
		vo.CpuFrequencyGhz, vo.MemoryTotalBytes.StringWithSuffix(),
	)
}
