package deltabackend.domain.sequenceDelta;

import java.util.List;

public class SequenceDDMinResponse {

    public SequenceDDMinResponse(){

    }

    private boolean status;
    private String message;
    private List<String> ddminResult;

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

    public List<String> getDdminResult() {
        return ddminResult;
    }

    public void setDdminResult(List<String> ddminResult) {
        this.ddminResult = ddminResult;
    }
}
