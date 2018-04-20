package deltabackend.domain.sequenceDelta;

import deltabackend.domain.api.request.SetAsyncRequestSequenceRequestWithSource;
import deltabackend.domain.test.DeltaTestResponse;

import java.util.List;

public class SequenceDeltaResponse {

    boolean status;
    String message;
    List<SetAsyncRequestSequenceRequestWithSource> envList;
    DeltaTestResponse result;
    boolean diffFromFirst;//different from the first test result, highlight it

    public SequenceDeltaResponse(){

    }

    public List<SetAsyncRequestSequenceRequestWithSource> getEnvList() {
        return envList;
    }

    public void setEnvList(List<SetAsyncRequestSequenceRequestWithSource> envList) {
        this.envList = envList;
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



    public DeltaTestResponse getResult() {
        return result;
    }

    public void setResult(DeltaTestResponse result) {
        this.result = result;
    }

    public boolean isDiffFromFirst() {
        return diffFromFirst;
    }

    public void setDiffFromFirst(boolean diffFromFirst) {
        this.diffFromFirst = diffFromFirst;
    }


}
