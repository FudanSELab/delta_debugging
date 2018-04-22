package deltabackend.domain.sequenceDelta;

import java.util.ArrayList;
import java.util.List;

public class SequenceDeltaRequest {

    String id;
    List<SingleSequenceDelta> seqGroups;
    List<String> tests;


    public List<SingleSequenceDelta> getSeqGroups() {
        return seqGroups;
    }

    public void setSeqGroups(List<SingleSequenceDelta> seqGroups) {
        this.seqGroups = seqGroups;
    }


    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }



    public List<String> getTests() {
        return tests;
    }

    public void setTests(List<String> tests) {
        this.tests = tests;
    }

    public SequenceDeltaRequest(){

    }


}
