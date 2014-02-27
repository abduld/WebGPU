package stats

import (
	"github.com/abduld/nvml-go"
	"strconv"
)

func gpuDevicecount() (int, error) {
	return nvml.DeviceCount()
}

func gpuDeviceNames() ([]string, error) {
	var names []string
	count, err := nvml.DeviceCount()
	if err != nil {
		return names, err
	}
	for ii := 0; ii < count; ii++ {
		handle, err := nvml.DeviceGetHandleByIndex(ii)
		if err != nil {
			return names, err
		}
		name, err := nvml.DeviceName(handle)
		if err != nil {
			return names, err
		}
		names = append(names, name)
	}
	return names, nil
}

func gpuIndexedDeviceName() ([]string, error) {
	names, err := gpuDeviceNames()
	if err == nil {
		for ii, name := range names {
			names[ii] = name + "(" + strconv.Itoa(ii) + ")"
		}
	}
	return names, err
}

func gpuTemperatures() ([]uint, error) {
	var temps []uint
	count, err := nvml.DeviceCount()
	if err != nil {
		return temps, err
	}
	for ii := 0; ii < count; ii++ {
		handle, err := nvml.DeviceGetHandleByIndex(ii)
		if err != nil {
			return temps, err
		}
		tmp, err := nvml.DeviceTemperature(handle)
		if err != nil {
			return temps, err
		}
		temps = append(temps, tmp)
	}
	return temps, nil
}

func gpuFanSpeeds() ([]uint, error) {
	var fans []uint
	count, err := nvml.DeviceCount()
	if err != nil {
		return fans, err
	}
	for ii := 0; ii < count; ii++ {
		handle, err := nvml.DeviceGetHandleByIndex(ii)
		if err != nil {
			return fans, err
		}
		fan, err := nvml.DeviceFanSpeed(handle)
		if err != nil {
			return fans, err
		}
		fans = append(fans, fan)
	}
	return fans, nil
}

func gpuMemoryInformation() ([]nvml.MemoryInformation, error) {
	var free []nvml.MemoryInformation
	count, err := nvml.DeviceCount()
	if err != nil {
		return free, err
	}
	for ii := 0; ii < count; ii++ {
		handle, err := nvml.DeviceGetHandleByIndex(ii)
		if err != nil {
			return free, err
		}
		mem, err := nvml.DeviceMemoryInformation(handle)
		if err != nil {
			return free, err
		}
		free = append(free, mem)
	}
	return free, nil
}

func gpuMemoryUsed() ([]uint64, error) {
	info, err := gpuMemoryInformation()
	if err == nil {
		used := make([]uint64, len(info))
		for ii, val := range info {
			used[ii] = val.Used
		}
		return used, nil
	}
	return nil, err
}

func gpuMemoryFree() ([]uint64, error) {
	info, err := gpuMemoryInformation()
	if err == nil {
		free := make([]uint64, len(info))
		for ii, val := range info {
			free[ii] = val.Free
		}
		return free, nil
	}
	return nil, err
}

func gpuMemoryTotal() ([]uint64, error) {
	info, err := gpuMemoryInformation()
	if err == nil {
		total := make([]uint64, len(info))
		for ii, val := range info {
			total[ii] = val.Total
		}
		return total, nil
	}
	return nil, err
}
