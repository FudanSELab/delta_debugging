package deltabackend.domain.serviceDelta;

public class RestartServiceResponse {
    private boolean status;
    private String message;

    public RestartServiceResponse(){

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
