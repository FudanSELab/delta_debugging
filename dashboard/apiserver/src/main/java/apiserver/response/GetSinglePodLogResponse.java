package apiserver.response;

import apiserver.bean.PodLog;

public class GetSinglePodLogResponse {
    private boolean status;
    private String message;
    private PodLog podLog;

    public GetSinglePodLogResponse(){

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

    public PodLog getPodLog() {
        return podLog;
    }

    public void setPodLog(PodLog podLog) {
        this.podLog = podLog;
    }
}
