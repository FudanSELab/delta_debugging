package deltabackend.domain.mixerDelta;

import deltabackend.domain.bean.SingleDeltaCMResourceRequest;
import deltabackend.domain.sequenceDelta.SingleSequenceDelta;

import java.util.ArrayList;
import java.util.List;

public class MixerDeltaRequest {

    String id;
    List<String> instances;
    List<SingleSequenceDelta> seqGroups;
    List<SingleDeltaCMResourceRequest> configs;
    List<String> tests;

    public MixerDeltaRequest(){

    }


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

    public List<String> getInstances() {
        return instances;
    }

    public void setInstances(List<String> instances) {
        this.instances = instances;
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
