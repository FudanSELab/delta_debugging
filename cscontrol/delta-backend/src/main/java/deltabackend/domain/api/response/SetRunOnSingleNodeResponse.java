package deltabackend.domain.api.response;

public class SetRunOnSingleNodeResponse {
    private boolean status;
    private String message;

    public SetRunOnSingleNodeResponse(){

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
