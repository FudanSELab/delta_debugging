package apiserver.response;

import apiserver.bean.PodLog;

import java.util.List;

public class GetPodsLogResponse {
    private boolean status;
    private String message;
    private List<PodLog> podLogs;

    public GetPodsLogResponse(){

    }

    public boolean isStatus() {
        return status;
    }

    public void setStatus(boolean status) {
        this.status = status;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }

    public List<PodLog> getPodLogs() {
        return podLogs;
    }

    public void setPodLogs(List<PodLog> podLogs) {
        this.podLogs = podLogs;
    }
}
