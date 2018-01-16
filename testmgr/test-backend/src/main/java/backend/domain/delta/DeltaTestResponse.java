package backend.domain.delta;

import java.util.ArrayList;
import java.util.List;

public class DeltaTestResponse {

    private boolean status;
    private String message;
    List<DeltaTestResult> deltaResults = new ArrayList<DeltaTestResult>();

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
