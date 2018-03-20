package apiserver.bean;

public class V1ContainerStateWaiting {
    private String message = null;
    private String reason = null;

    public V1ContainerStateWaiting(){

    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }

    public String getReason() {
        return reason;
    }

    public void setReason(String reason) {
        this.reason = reason;
    }
}
