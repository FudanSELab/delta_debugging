package deltabackend.domain.mixerDelta;

import deltabackend.domain.bean.SingleDeltaCMResourceRequest;

import java.util.ArrayList;
import java.util.List;

public class MixerDeltaRequest {

    String id;
    List<String> instances;
    String sender;
    ArrayList<String> receivers;
    List<SingleDeltaCMResourceRequest> configs;
    List<String> tests;

    public MixerDeltaRequest(){

    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public List<String> getInstances() {
        return instances;
    }

    public void setInstances(List<String> instances) {
        this.instances = instances;
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

    public List<SingleDeltaCMResourceRequest> getConfigs() {
        return configs;
    }

    public void setConfigs(List<SingleDeltaCMResourceRequest> configs) {
        this.configs = configs;
    }

    public List<String> getTests() {
        return tests;
    }

    public void setTests(List<String> tests) {
        this.tests = tests;
    }


}
