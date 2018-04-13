package deltabackend.domain.sequenceDelta;

import deltabackend.domain.DeltaTestResponse;

import java.util.ArrayList;
import java.util.List;

public class SequenceDeltaResponse {

    boolean status;
    String message;
    String sender;
    List<String> receiversInOrder = null;
    DeltaTestResponse result;
    boolean diffFromFirst;//different from the first test result, highlight it

    public SequenceDeltaResponse(){

    }

    public List<String> getReceiversInOrder() {
        return receiversInOrder;
    }

    public void setReceiversInOrder(List<String> receiversInOrder) {
        this.receiversInOrder = receiversInOrder;
    }

    public String getSender() {
        return sender;
    }

    public void setSender(String sender) {
        this.sender = sender;
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
