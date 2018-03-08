package apiserver.bean;

public class V1NodeSystemInfo {
    private String architecture = null;
    private String bootID = null;
    private String containerRuntimeVersion = null;
    private String kernelVersion = null;
    private String kubeProxyVersion = null;
    private String kubeletVersion = null;
    private String machineID = null;
    private String operatingSystem = null;
    private String osImage = null;
    private String systemUUID = null;

    public V1NodeSystemInfo(){

    }

    public String getArchitecture() {
        return architecture;
    }

    public void setArchitecture(String architecture) {
        this.architecture = architecture;
    }

    public String getBootID() {
        return bootID;
    }

    public void setBootID(String bootID) {
        this.bootID = bootID;
    }

    public String getContainerRuntimeVersion() {
        return containerRuntimeVersion;
    }

    public void setContainerRuntimeVersion(String containerRuntimeVersion) {
        this.containerRuntimeVersion = containerRuntimeVersion;
    }

    public String getKernelVersion() {
        return kernelVersion;
    }

    public void setKernelVersion(String kernelVersion) {
        this.kernelVersion = kernelVersion;
    }

    public String getKubeProxyVersion() {
        return kubeProxyVersion;
    }

    public void setKubeProxyVersion(String kubeProxyVersion) {
        this.kubeProxyVersion = kubeProxyVersion;
    }

    public String getKubeletVersion() {
        return kubeletVersion;
    }

    public void setKubeletVersion(String kubeletVersion) {
        this.kubeletVersion = kubeletVersion;
    }

    public String getMachineID() {
        return machineID;
    }

    public void setMachineID(String machineID) {
        this.machineID = machineID;
    }

    public String getOperatingSystem() {
        return operatingSystem;
    }

    public void setOperatingSystem(String operatingSystem) {
        this.operatingSystem = operatingSystem;
    }

    public String getOsImage() {
        return osImage;
    }

    public void setOsImage(String osImage) {
        this.osImage = osImage;
    }

    public String getSystemUUID() {
        return systemUUID;
    }

    public void setSystemUUID(String systemUUID) {
        this.systemUUID = systemUUID;
    }
}
