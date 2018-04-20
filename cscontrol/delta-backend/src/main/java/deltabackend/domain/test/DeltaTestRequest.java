package deltabackend.domain.test;

import java.util.List;

public class DeltaTestRequest {

    String cluster;
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
