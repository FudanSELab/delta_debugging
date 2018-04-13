package deltabackend.domain.sequenceDelta;

public class SetAsyncRequestSequenceResponse {

    private boolean status;

    private String message;

    public SetAsyncRequestSequenceResponse() {
        //do nothing
    }

    public SetAsyncRequestSequenceResponse(boolean status, String message) {
        this.status = status;
        this.message = message;
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
