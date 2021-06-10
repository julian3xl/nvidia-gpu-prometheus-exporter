package main

import (
    "io"
    "net/http"
    "encoding/xml"
    "os/exec"
    "log"
    "regexp"
    "strconv"
)

const LISTEN_ADDRESS = ":9202"
const NVIDIA_SMI_PATH = "/usr/bin/nvidia-smi"

type NvidiaSmiLog struct {
    DriverVersion string `xml:"driver_version"`
    CudaVersion string `xml:"cuda_version"`
    AttachedGPUs string `xml:"attached_gpus"`
    GPUs []struct {
        ProductName string `xml:"product_name"`
        ProductBrand string `xml:"product_brand"`
        UUID string `xml:"uuid"`
        FanSpeed string `xml:"fan_speed"`
        PerformanceState string `xml:"performance_state"`
        PCI struct {
            PCIBus string `xml:"pci_bus"`
        } `xml:"pci"`
        FbMemoryUsage struct {
            Total string `xml:"total"`
            Used string `xml:"used"`
            Free string `xml:"free"`
        } `xml:"fb_memory_usage"`
        Utilization struct {
            GPUUtil string `xml:"gpu_util"`
            MemoryUtil string `xml:"memory_util"`
        } `xml:"utilization"`
        Temperature struct {
            GPUTemp string `xml:"gpu_temp"`
            GPUTempMaxThreshold string `xml:"gpu_temp_max_threshold"`
            GPUTempSlowThreshold string `xml:"gpu_temp_slow_threshold"`
        } `xml:"temperature"`
        PowerReadings struct {
            PowerState string `xml:"power_state"`
            PowerDraw string `xml:"power_draw"`
            PowerLimit string `xml:"power_limit"`
            DefaultPowerLimit string `xml:"default_power_limit"`
            EnforcedPowerLimit string `xml:"enforced_power_limit"`
            MinPowerLimit string `xml:"min_power_limit"`
            MaxPowerLimit string `xml:"max_power_limit"`
        } `xml:"power_readings"`
        Clocks struct {
            GraphicsClock string `xml:"graphics_clock"`
            SmClock string `xml:"sm_clock"`
            MemClock string `xml:"mem_clock"`
            VideoClock string `xml:"video_clock"`
        } `xml:"clocks"`
        MaxClocks struct {
            GraphicsClock string `xml:"graphics_clock"`
            SmClock string `xml:"sm_clock"`
            MemClock string `xml:"mem_clock"`
            VideoClock string `xml:"video_clock"`
        } `xml:"max_clocks"`
        Processes struct {
            ProcessInfo []struct {
                Pid string `xml:"pid"`
                ProcessName string `xml:"process_name"`
                UsedMemory string `xml:"used_memory"`
            } `xml:"process_info"`
        } `xml:"processes"`
    } `xml:"gpu"`
}

func formatValue(key string, meta string, value string) string {
    result := key;
    if (meta != "") {
        result += "{" + meta + "}";
    }
    return result + " " + value +"\n"
}

func filterNumber(value string) string {
    r := regexp.MustCompile("[^0-9.]")
    return r.ReplaceAllString(value, "")
}

func metrics(w http.ResponseWriter, r *http.Request) {
    log.Print("Serving /metrics")

    var cmd *exec.Cmd
    cmd = exec.Command(NVIDIA_SMI_PATH, "-q", "-x")

    // Execute system command
    stdout, err := cmd.Output()
    if err != nil {
        println(err.Error())
        return
    }

    // Parse XML
    var xmlData NvidiaSmiLog
    xml.Unmarshal(stdout, &xmlData)

    // Output
    var gpuMetadata string
    io.WriteString(w, formatValue("nvidia_attached_gpus", "", filterNumber(xmlData.AttachedGPUs)))
    io.WriteString(w, formatValue("nvidia_cuda_version", "", xmlData.CudaVersion))
    io.WriteString(w, formatValue("nvidia_driver_version", "", xmlData.DriverVersion))
    for i, GPU := range xmlData.GPUs {
        gpuMetadata = "gpu=\"" + strconv.Itoa(i) + "\", product_name=\"" + GPU.ProductName + "\", uuid=\"" + GPU.UUID + "\""

        io.WriteString(w, formatValue("nvidia_clock_graphics_max", gpuMetadata, filterNumber(GPU.MaxClocks.GraphicsClock)))
        io.WriteString(w, formatValue("nvidia_clock_graphics", gpuMetadata, filterNumber(GPU.Clocks.GraphicsClock)))
        io.WriteString(w, formatValue("nvidia_clock_mem_max", gpuMetadata, filterNumber(GPU.MaxClocks.MemClock)))
        io.WriteString(w, formatValue("nvidia_clock_mem", gpuMetadata, filterNumber(GPU.Clocks.MemClock)))
        io.WriteString(w, formatValue("nvidia_clock_sm_max", gpuMetadata, filterNumber(GPU.MaxClocks.SmClock)))
        io.WriteString(w, formatValue("nvidia_clock_sm", gpuMetadata, filterNumber(GPU.Clocks.SmClock)))
        io.WriteString(w, formatValue("nvidia_clock_video_max", gpuMetadata, filterNumber(GPU.MaxClocks.VideoClock)))
        io.WriteString(w, formatValue("nvidia_clock_video", gpuMetadata, filterNumber(GPU.Clocks.VideoClock)))
        io.WriteString(w, formatValue("nvidia_fan_speed", gpuMetadata, filterNumber(GPU.FanSpeed)))
        io.WriteString(w, formatValue("nvidia_memory_usage_free", gpuMetadata, filterNumber(GPU.FbMemoryUsage.Free)))
        io.WriteString(w, formatValue("nvidia_memory_usage_total", gpuMetadata, filterNumber(GPU.FbMemoryUsage.Total)))
        io.WriteString(w, formatValue("nvidia_memory_usage_used", gpuMetadata, filterNumber(GPU.FbMemoryUsage.Used)))
        io.WriteString(w, formatValue("nvidia_power_draw", gpuMetadata, filterNumber(GPU.PowerReadings.PowerDraw)))
        io.WriteString(w, formatValue("nvidia_power_limit", gpuMetadata, filterNumber(GPU.PowerReadings.PowerLimit)))
        io.WriteString(w, formatValue("nvidia_gpu_temp_max_threshold", gpuMetadata, filterNumber(GPU.Temperature.GPUTempMaxThreshold)))
        io.WriteString(w, formatValue("nvidia_gpu_temp_slow_threshold", gpuMetadata, filterNumber(GPU.Temperature.GPUTempSlowThreshold)))
        io.WriteString(w, formatValue("nvidia_gpu_temp", gpuMetadata, filterNumber(GPU.Temperature.GPUTemp)))
        io.WriteString(w, formatValue("nvidia_utilization_gpu", gpuMetadata, filterNumber(GPU.Utilization.GPUUtil)))
        io.WriteString(w, formatValue("nvidia_utilization_memory", gpuMetadata, filterNumber(GPU.Utilization.MemoryUtil)))
        io.WriteString(w, formatValue("nvidia_performance_state", gpuMetadata, filterNumber(GPU.PerformanceState)))
        io.WriteString(w, formatValue("nvidia_power_state", gpuMetadata, filterNumber(GPU.PowerReadings.PowerState)))
        io.WriteString(w, formatValue("nvidia_default_power_limit", gpuMetadata, filterNumber(GPU.PowerReadings.DefaultPowerLimit)))
        io.WriteString(w, formatValue("nvidia_enforced_power_limit", gpuMetadata, filterNumber(GPU.PowerReadings.EnforcedPowerLimit)))
        io.WriteString(w, formatValue("nvidia_min_power_limit", gpuMetadata, filterNumber(GPU.PowerReadings.MinPowerLimit)))
        io.WriteString(w, formatValue("nvidia_max_power_limit", gpuMetadata, filterNumber(GPU.PowerReadings.MaxPowerLimit)))

        for j, ProcessInfo := range GPU.Processes.ProcessInfo {
            io.WriteString(w, formatValue("nvidia_process_info_pid", gpuMetadata + ", process=\"" + strconv.Itoa(j) + "\"", filterNumber(ProcessInfo.Pid)))
            io.WriteString(w, formatValue("nvidia_process_info_name", gpuMetadata + ", process=\"" + strconv.Itoa(j) + "\"", filterNumber(ProcessInfo.ProcessName)))
            io.WriteString(w, formatValue("nvidia_process_info_used_memory", gpuMetadata + ", process=\"" + strconv.Itoa(j) + "\"", filterNumber(ProcessInfo.UsedMemory)))
        }
    }
}

func index(w http.ResponseWriter, r *http.Request) {
    log.Print("Serving /index")
    html := `<!doctype html>
<html>
    <head>
        <meta charset="utf-8">
        <title>Nvidia SMI Exporter</title>
    </head>
    <body>
        <h1>Nvidia SMI Exporter</h1>
        <p><a href="/metrics">Metrics</a></p>
    </body>
</html>`
    io.WriteString(w, html)
}

func main() {
    log.Print("Nvidia SMI exporter listening on " + LISTEN_ADDRESS)
    http.HandleFunc("/", index)
    http.HandleFunc("/metrics", metrics)
    http.ListenAndServe(LISTEN_ADDRESS, nil)
}
