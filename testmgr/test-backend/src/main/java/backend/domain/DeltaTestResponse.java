package backend.domain;

import java.util.ArrayList;
import java.util.List;

public class DeltaTestResponse {

    private boolean status;
    private String message;
    List<TestResponse> deltaResults = new ArrayList<TestResponse>();

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

    public List<TestResponse> getDeltaResults() {
        return deltaResults;
    }

    public void setDeltaResults(List<TestResponse> deltaResults) {
        this.deltaResults = deltaResults;
    }

    public void addDeltaResult(TestResponse t){
        this.deltaResults.add(t);
    }



}
