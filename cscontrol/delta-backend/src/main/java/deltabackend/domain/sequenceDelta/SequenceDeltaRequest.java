package deltabackend.domain.sequenceDelta;

import java.util.ArrayList;
import java.util.List;

public class SequenceDeltaRequest {

    String id;

    String sender;

    ArrayList<String> receivers;

    List<String> tests;


    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getSender() {
        return sender;
    }

    public void setSender(String sender) {
        this.sender = sender;
    }

    public ArrayList<String> getReceivers() {
        return receivers;
    }

    public void setReceivers(ArrayList<String> receivers) {
        this.receivers = receivers;
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
