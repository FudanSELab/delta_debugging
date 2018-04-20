package deltabackend.domain.serviceDelta;

import java.util.List;

public class ServiceDeltaRequest {

    String id;
    List<String> tests;
    String cluster;

    public String getCluster() {
        return cluster;
    }

    public void setCluster(String cluster) {
        this.cluster = cluster;
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
}
