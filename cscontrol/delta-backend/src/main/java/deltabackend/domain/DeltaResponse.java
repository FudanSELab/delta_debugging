package deltabackend.domain;

import java.util.ArrayList;
import java.util.List;

public class DeltaResponse {
    boolean status;
    String message;
    List<EnvParameter> env = new ArrayList<EnvParameter>();
    DeltaTestResponse result;
    boolean diffFromFirst;//different from the first test result, highlight it

    public DeltaTestResponse getResult() {
        return result;
    }

    public void setResult(DeltaTestResponse result) {
        this.result = result;
    }

    public List<EnvParameter> getEnv() {
        return env;
    }

    public void setEnv(List<EnvParameter> env) {
        this.env = env;
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

    public boolean isDiffFromFirst() {
        return diffFromFirst;
    }

    public void setDiffFromFirst(boolean diffFromFirst) {
        this.diffFromFirst = diffFromFirst;
    }


}
