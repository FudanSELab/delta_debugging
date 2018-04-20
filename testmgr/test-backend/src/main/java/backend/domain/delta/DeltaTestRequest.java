package backend.domain.delta;

import java.util.List;

public class DeltaTestRequest {

    String cluster = null;

    List<String> testNames;

    public String getCluster() {
        return cluster;
    }

    public void setCluster(String cluster) {
        this.cluster = cluster;
    }

    public List<String> getTestNames() {
        return testNames;
    }

    public void setTestNames(List<String> testNames) {
        this.testNames = testNames;
    }


}
