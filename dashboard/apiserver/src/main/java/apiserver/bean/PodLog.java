package apiserver.bean;

public class PodLog {
    private String podName;
    private String logs;

    public PodLog(){

    }

    public String getPodName() {
        return podName;
    }

    public void setPodName(String podName) {
        this.podName = podName;
    }

    public String getLogs() {
        return logs;
    }

    public void setLogs(String logs) {
        this.logs = logs;
    }
}
