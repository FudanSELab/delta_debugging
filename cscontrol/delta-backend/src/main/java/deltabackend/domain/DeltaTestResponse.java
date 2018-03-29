package deltabackend.domain;

import java.util.ArrayList;
import java.util.List;

public class DeltaTestResponse {

    private int status;//0:FAUILUE; 1:SUCCESS; -1:EXCEPTION
    private String message;
    List<DeltaTestResult> deltaResults = new ArrayList<DeltaTestResult>();

    public int getStatus() {
        return status;
    }

    public void setStatus(int status) {
        this.status = status;
    }


    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }


    public List<DeltaTestResult> getDeltaResults() {
        return deltaResults;
    }

    public void setDeltaResults(List<DeltaTestResult> result) {
        this.deltaResults = result;
    }

    public void addDeltaResult(DeltaTestResult d){
        deltaResults.add(d);
    }


}
