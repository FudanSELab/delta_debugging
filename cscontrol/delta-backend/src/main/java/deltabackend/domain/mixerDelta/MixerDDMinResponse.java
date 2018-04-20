package deltabackend.domain.mixerDelta;

import java.util.List;
import java.util.Map;

public class MixerDDMinResponse {
    private boolean status;
    private String message;
    private Map<String, List<String>> ddminResult;

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

    public Map<String, List<String>> getDdminResult() {
        return ddminResult;
    }

    public void setDdminResult(Map<String, List<String>> ddminResult) {
        this.ddminResult = ddminResult;
    }

}
