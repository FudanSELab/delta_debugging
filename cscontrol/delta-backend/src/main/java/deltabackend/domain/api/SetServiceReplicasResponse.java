package deltabackend.domain.api;

public class SetServiceReplicasResponse {
    private boolean status;
    private String message;

    public SetServiceReplicasResponse(){

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
}
